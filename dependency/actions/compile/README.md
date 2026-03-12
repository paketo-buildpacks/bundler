Running compilation locally:

1. Build the build environment:
```
docker build --platform <os>/<arch> -t compilation -f <target>.Dockerfile dependency/actions/compile
```

2. Make the output directory:
```
mkdir <output dir>
```

3. Run compilation and use a volume mount to access it:
```
docker run --platform <os>/<arch> -v <output dir>:/output --rm compilation --version <version> --outputDir /output --target <target> --os <os> --arch <arch>
```

Notes:
- <target> can be: jammy or noble
- <os>: linux
- <arch>: amd64 or arm64
- If you omit --platform/--os/--arch, defaults are linux/amd64.
- Bundler 2.7.x requires Ruby 3.2+, so the compile/test containers use an Ubuntu 24.04 base even for the jammy target.
- The legacy ubuntu target has been removed. Use jammy for Ubuntu 22.04-compatible artifacts.

Example for Bundler 2.5.18 on noble/arm64:
```
docker build --platform linux/arm64 -t compilation -f noble.Dockerfile dependency/actions/compile
docker run --platform linux/arm64 -v ~/bundler-build:/output --rm compilation --version 2.5.18 --outputDir /output --target noble --os linux --arch arm64
```

Example for Bundler 2.5.18 on jammy/amd64:
```
docker build --platform linux/amd64 -t compilation -f jammy.Dockerfile dependency/actions/compile
docker run --platform linux/amd64 -v ~/bundler-build:/output --rm compilation --version 2.5.18 --outputDir /output --target jammy --os linux --arch amd64
```
