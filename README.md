# Bundler Cloud Native Buildpack

## `gcr.io/paketo-buildpacks/bundler`

The Bundler CNB provides the Bundler binary. The buildpack installs the Bundler 
onto the `$PATH` and `$GEM_PATH` which makes it available for subsequent buildpacks
and in the final running container.

## Integration

The Bundler CNB provides bundler as a dependency. Downstream buildpacks, can require the bundler dependency
by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the Bundler dependency is "bundler". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "bundler"

  # The version of the Bundler dependency is not required. In the case it
  # is not specified, the buildpack will provide the default version, which can
  # be seen in the buildpack.toml file.
  # If you wish to request a specific version, the buildpack supports
  # specifying a semver constraint in the form of "1.*", "1.17.*", or even
  # "1.17.1".
  version = "1.17.1"

  # The Bundler buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the Bundler
    # depdendency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to run Bundle
    # during its build process, this flag should be set to true.
    build = true
```
## Usage

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## `buildpack.yml` Configurations

```yaml
bundler:
  # this allows you to specify a version constaint for the Bundler dependency
  # any valid semver constaints (e.g. 2.* and 2.1.*) are also acceptable
  version: "2.1.4"
```
