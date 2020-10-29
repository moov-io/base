# MySQL docker image for testing

In order to speedup testing with MySQL we created custom docker image for MySQL that stores
data inside container (not outside in volume).

# Update Image

In order to update image please follow this steps:

Updated `Dockerfile` if needed and then build the image:

```
make build
```

Run the container from this image to initialize database and create all necessary files inside container

```
make run
```

After container is ready (you will see that it's "...ready for connections..."), please, stop the container with

```
make stop
```

at this point we have a container with initialized MySQL database

Now it's time to make image based on this container:

```
make image
```

Final step is to push image to docker hub:

```
make push
```

That's it!
