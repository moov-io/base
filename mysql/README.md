# MySQL docker image for testing

In order to speedup testing with MySQL we created custom docker image for MySQL that stores
data inside container (not outside in volume).

The image is based on `mysql:26.10-oracle` and is tagged `mysql-volumeless:26.10` (and `moov/mysql-volumeless:26.10`).

# Update Image

Updated `Dockerfile` if needed and then build, initialize and commit the image:

```
make build
```

`make build` builds the image, starts a container, waits until MySQL is ready, stops the container and commits the result as `mysql-volumeless:26.10` and `moov/mysql-volumeless:26.10`.

Final step is to push image to docker hub:

```
make push
```

That's it!
