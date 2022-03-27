# 0xC1

> Becuase we all know that we don't have enough tools for devops.

**0xC1** aim to provide a simple and easy to use CI/CD solution, that stays together with the unix philosophy, it is supposed to _Do One Thing and Do It Well_.

That being said, let's figure out if 0xC1 is something for you. Here are some things worth mentioning.

 - One central configuration for all projects.
 - Everything is being built in docker containers.
 - Everything is a shell script.
 - Non-complicated configuration.
 - Fully featured CI/CD

Even thought the `status` of a job is limited to 'success' and 'fail', logging is in fact possible so you can use any tool you want to run test, save them in file and then read them from your browser.

### Details

Instructions for a CI instance are stored in a git repository.
There are several options for using our ci, example configuration can be found in `_example` directory.

https://github.com/docker/buildx/issues/132#issuecomment-532350357
https://github.com/docker/buildx/issues/132#issuecomment-582218096

Generally use docker following the docs.