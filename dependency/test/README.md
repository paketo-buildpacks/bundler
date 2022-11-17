### Running tests locally:
1. From this directory's parent, run
  ```
  make test tarballPath="path/to/compiled-bundler.tgz" version="1.2.3"
  ```
2. The make target will build Docker containers, mount the artifact into them,
   and run a series of tests.
3. If the make target commpletes without error, the tests have passed.
