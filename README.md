# gitlab2gogs

Migrate your GitLab 9.x repositories to Gogs.

**Build Status:** [![Build Status](https://travis-ci.org/CHERTS/gitlab2gogs.svg?branch=master)](https://travis-ci.org/CHERTS/gitlab2gogs)

### Usage

```
./gitlab2gogs -gitlab-host https://<yourgitlabhost> \
    -gitlab-api-path /api/v4
    -gitlab-token <your gitlab token> \
    -gitlab-user <gitlab admin user> \
    -gitlab-password <password of gitlab-user> \
    -gitlab-visibilitylevel {private|internal|public} \
    -gitlab-repo <repository name (optional)> \
    -gitlab-org <organization name (optional)> \
    -gogs-url https://<yourgogshost> \
    -gogs-token <your gogs token> \
    -gogs-user <gogs admin username>
```

### Features

Organizations are created if they do not yet exists.

Users are created if they do not yet exists.

Existing repositories (in Gogs) are not overwritten.

For migration of Repositories in a single Organization: `-gitlab-org <organization name>`

And to migrate a single Repository within that Organization: `-gitlab-org <organization name> -gitlab-repo <repository name>`

Or to migrate a single Repository without Organization: `-gitlab-repo <repository name>`

To make migrated Repositories as mirror (backup) Repository: `-mirror`

### Build on Debian Linux

#### Install GoLang

```
wget https://storage.googleapis.com/golang/go1.8.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.8.linux-amd64.tar.gz
mkdir /usr/local/go/packages
echo 'export GOROOT=/usr/local/go'>/etc/profile.d/golang.sh
echo 'export GOPATH=/usr/local/go/packages'>>/etc/profile.d/golang.sh
echo 'export PATH=$GOROOT/bin:$GOPATH/bin:$PATH'>>/etc/profile.d/golang.sh
echo 'export GOROOT=/usr/local/go'>>~/.bashrc
echo 'export GOPATH=/usr/local/go/packages'>>~/.bashrc
echo 'export PATH=$GOROOT/bin:$GOPATH/bin:$PATH'>>~/.bashrc
source ~/.bashrc
```

#### Checking GoLang version

```
# go version
go version go1.8 linux/amd64
```

#### Build gitlab2gogs

```
git clone https://github.com/CHERTS/gitlab2gogs
cd gitlab2gogs
make
```
