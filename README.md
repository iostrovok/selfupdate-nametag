# nametag

A program that updates itself.

Imagine you have a program that we deployed to customers and periodically release new versions. 
When a new version is released, we want the deployed programs to be seamlessly replaced with the new version.

How to run the tests:

Install go if you already have it.

open the first console and run the following commands:

```bash
make install
make run version="v0.0.1"
```

Check the version in your browser: http://127.0.0.1:8081

Open the second console. We are going to start a server that will serve versions and images.

```bash
make run-images-server
```

Open the third console. Here we will create new images with new versions.
After each command you can check the version in your browser. It can take up to 20 seconds to update the version.
The next version should be larger than the previous one.

See more about versions: https://semver.org/

```bash
make build version="v0.0.2"
...
make build version="v0.0.3"
...
make build version="v0.1.7"
...
make build version="v4.0.0"
...
```


The test server is not optimized to run in console mode.
Therefore, if you want to permanently stop your test server, 
you can use the following command and kill the process.

```bash
ps awx | grep exe/app
```