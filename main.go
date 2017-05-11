package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gogits/go-gogs-client"
	"github.com/xanzy/go-gitlab"
)

var (
	gitlabHost            string
	gitlabApiPath         string
	gitlabUser            string
	gitlabPassword        string
	gitlabToken           string
	gitlabVisibilityLevel string
	gogsUrl               string
	gogsToken             string
	gogsUser              string
)

func init() {
	flag.StringVar(&gitlabHost, "gitlab-host", "https://gitlab", "GitLab URL address")
	flag.StringVar(&gitlabApiPath, "gitlab-api-path", "/api/v4", "GitLab API URL")
	flag.StringVar(&gitlabUser, "gitlab-user", "root", "GitLab user")
	flag.StringVar(&gitlabPassword, "gitlab-password", "", "GitLab user password")
	flag.StringVar(&gitlabToken, "gitlab-token", "", "GitLab user access token")
	flag.StringVar(&gitlabVisibilityLevel, "gitlab-visibilitylevel", "private", "GitLab repositary visibility level (private, internal, public)")
	flag.StringVar(&gogsUrl, "gogs-url", "https://gogs", "Gogs URL address")
	flag.StringVar(&gogsToken, "gogs-token", "", "Gogs user access token")
	flag.StringVar(&gogsUser, "gogs-user", "root", "Gogs user")
}

func main() {
	flag.Parse()

	git := gitlab.NewClient(nil, gitlabToken)
	git.SetBaseURL(gitlabHost + gitlabApiPath)

	gc := gogs.NewClient(gogsUrl, gogsToken)
	orgMap := make(map[string]*gogs.Organization)
	userMap := make(map[string]*gogs.User)
	gitlabuserMap := make(map[string]*gitlab.User)
	gitlabgroupMap := make(map[string]*gitlab.Group)

	getGogsOrg := func(gitlaborg *gitlab.Group) *gogs.Organization {
		name := fixName(gitlaborg.Name)
		org, ok := orgMap[name]
		if ok {
			return org
		}
		org, err := gc.GetOrg(name)
		if err == nil {
			orgMap[name] = org
			return org
		}
		createOpt := gogs.CreateOrgOption{
			UserName:    name,
			FullName:    gitlaborg.Name,
			Description: gitlaborg.Description,
		}
		org, err = gc.AdminCreateOrg(gogsUser, createOpt)
		if err != nil {
			exitf("Failed to create organization '%s': %v\n", name, err)
		}
		orgMap[name] = org
		return org
	}

	getGogsUser := func(gitlabuser *gitlab.User) *gogs.User {
		user, ok := userMap[gitlabuser.Username]
		if ok {
			return user
		}
		user, err := gc.GetUserInfo(gitlabuser.Username)
		if err == nil {
			userMap[gitlabuser.Username] = user
			return user
		}
		createOpt := gogs.CreateUserOption{
			Username: gitlabuser.Username,
			FullName: gitlabuser.Name,
			Email:    gitlabuser.Email,
		}
		user, err = gc.AdminCreateUser(createOpt)
		if err != nil {
			exitf("Failed to create user '%s': %v\n", gitlabuser.Username, err)
		}
		userMap[gitlabuser.Username] = user
		return user
	}

	getGitlabUser := func(owner *gitlab.User) *gitlab.User {
		gitlabuser, ok := gitlabuserMap[owner.Username]
		if ok {
			return gitlabuser
		}
		gitlabuser, _, err := git.Users.GetUser(owner.ID)
		if err != nil {
			exitf("Cannot get gitlab user: %v\n", err)
		}
		return gitlabuser
	}

	getGitlabGroup := func(gitlaborg *gitlab.ProjectNamespace) *gitlab.Group {
		gitlabgroup, ok := gitlabgroupMap[gitlaborg.Name]
		if ok {
			return gitlabgroup
		}
		gitlabgroup, _, err := git.Groups.GetGroup(gitlaborg.ID)
		if err != nil {
			exitf("Cannot get gitlab group: %v\n", err)
		}
		return gitlabgroup
	}

	migrate := func(p *gitlab.Project) {
		reponame := fixName(p.Name)
		owner := fixName(p.Namespace.Name)
		_, err := gc.GetRepo(owner, reponame)
		if err == nil {
			fmt.Printf("%s | %s already exists\n", owner, reponame)
		} else {
			if p.Owner != nil {
				gitlabuser := getGitlabUser(p.Owner)
				user := getGogsUser(gitlabuser)
				// Fix repo name
				name := fixName(p.Name)
				fmt.Printf("%s | %s migrating as '%s'... (GogsUser: ID: %d, UserName: %s, Email: %s)\n", p.Namespace.Name, p.Name, name, user.ID, user.UserName, user.Email)
				opts := gogs.MigrateRepoOption{
					CloneAddr:    p.HTTPURLToRepo,
					AuthUsername: gitlabUser,
					AuthPassword: gitlabPassword,
					UID:          int(user.ID),
					RepoName:     name,
					Private:      !p.Public,
					Description:  p.Description,
				}
				_, err := gc.MigrateRepo(opts)
				if err != nil {
					exitf("Failed to migrate '%s | %s': %v\n", p.Namespace.Name, p.Name, err)
				}
			} else {
				gitlabgroup := getGitlabGroup(p.Namespace)
				org := getGogsOrg(gitlabgroup)
				name := fixName(p.Name)
				fmt.Printf("%s | %s migrating as '%s'... (GogsOrg: ID:%d, FullName: %s, Description: %s)\n", p.Namespace.Name, p.Name, name, org.ID, org.FullName, org.Description)
				opts := gogs.MigrateRepoOption{
					CloneAddr:    p.HTTPURLToRepo,
					AuthUsername: gitlabUser,
					AuthPassword: gitlabPassword,
					UID:          int(org.ID),
					RepoName:     name,
					Private:      !p.Public,
					Description:  p.Description,
				}
				_, err := gc.MigrateRepo(opts)
				if err != nil {
					exitf("Failed to migrate '%s | %s': %v\n", p.Namespace.Name, p.Name, err)
				}
			}
		}
	}

	opt := &gitlab.ListProjectsOptions{Visibility: stringToVisibilityLevel(gitlabVisibilityLevel), OrderBy: gitlab.String("id"), Sort: gitlab.String("desc")}
	projects, _, err := git.Projects.ListProjects(opt)
	if err != nil {
		exitf("Cannot get gitlab projects: %v\n", err)
	}
	repo_cnt := 0
	for _, p := range projects {
		repo_cnt++
		migrate(p)
	}
	fmt.Printf("Total migrate repo: %d\n", repo_cnt)
}

func stringToVisibilityLevel(s string) *gitlab.VisibilityLevelValue {
	lookup := map[string]gitlab.VisibilityLevelValue{
		"private":  gitlab.PrivateVisibility,
		"internal": gitlab.InternalVisibility,
		"public":   gitlab.PublicVisibility,
	}

	value, ok := lookup[s]
	if !ok {
		return nil
	}
	return &value
}

func visibilityLevelToString(v gitlab.VisibilityLevelValue) *string {
	lookup := map[gitlab.VisibilityLevelValue]string{
		gitlab.PrivateVisibility:  "private",
		gitlab.InternalVisibility: "internal",
		gitlab.PublicVisibility:   "public",
	}
	value, ok := lookup[v]
	if !ok {
		return nil
	}
	return &value
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func fixName(name string) string {
	switch name {
	case "api": // reserved
		return "theapi"
	default:
		return name
	}
}

func exitf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Exit(1)
}
