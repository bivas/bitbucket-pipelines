# Examples

## Install

```
$ go install github.com/bivas/bitbucket-pipelines
```

## Usage

```
$ bitbucket-pipelines --help
Usage of bitbucket-pipelines:
  -env string
    	Add environment variables to pipeline
  -yaml string
    	Specify pipelines yaml file (default "bitbucket-pipelines.yml")
```

## Simple Run

Will run the local [`bitbucket-pipelines.yml`](bitbucket-pipelines.yml)


```
$ bitbucket-pipelines
2017/04/21 14:45:29 [bitbucket pipeline] Pulling image busybox
2017/04/21 14:45:31 [bitbucket pipeline] Running image busybox
2017/04/21 14:45:33 [bitbucket pipeline]  == Running 'ls' ==>
README.md
bitbucket-pipelines.yml
override.yml
vars.env
with-env.yaml
2017/04/21 14:45:33 [bitbucket pipeline]  == Running 'ps' ==>
PID   USER     TIME   COMMAND
    1 root       0:00 /bin/sh
    7 root       0:00 sleep 1
   13 root       0:00 /bin/sh -c ps
   18 root       0:00 ps

```

## Using a Different File

Will run the file [`override.yml`](override.yml)
```
$ bitbucket-pipelines --yaml override.yml
2017/04/21 14:45:29 [bitbucket pipeline] Pulling image busybox
2017/04/21 14:45:31 [bitbucket pipeline] Running image busybox
2017/04/21 14:45:33 [bitbucket pipeline]  == Running 'ls' ==>
README.md
bitbucket-pipelines.yml
override.yml
vars.env
with-env.yaml
```

## Specify an Environment for Pipeline

Will run the file [`with-env.yml`](with-env.yml) and set the environment to [`vars.env`](vars.env)

```
$ bitbucket-pipelines --yaml with-env.yml --env vars.env
2017/04/21 14:49:07 [bitbucket pipeline] Pulling image busybox
2017/04/21 14:49:10 [bitbucket pipeline] Running image busybox
2017/04/21 14:49:11 [bitbucket pipeline]  == Running 'env' ==>
no_proxy=*.local, 169.254/16
HOSTNAME=88448ca37b1a
SHLVL=1
HOME=/root
BAR=foo
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
FOO=bar
PWD=/wd
```