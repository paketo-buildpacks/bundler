Running tests locally:
1. Build the Dockerfile for testing (bionic or jammy)
```
docker build -f bionic.Dockerfile --tag test-bionic  .
```

2. Run the tests; mount the compiled artifact into the container and point the test script at it.
```
docker run -it -v /local/path/to/artifact/dir:/input test-bionic --artifact /input/name-of-artifact.tgz
```

3. If the container exits without error, the tests have passed.
