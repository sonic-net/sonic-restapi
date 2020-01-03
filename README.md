# SONiC-restapi

## Description
This is a configuration agent which exposes HTTPS endpoints to perform dynamic network configuration on switches running SONiC. It restful API server is `go-server-server`

## Getting Started
### Fetch Dependencies
  1. `mkdir debs; cd debs`
  2. `sudo wget https://sonic-jenkins.westus2.cloudapp.azure.com/job/vs/job/buildimage-vs-all/lastSuccessfulBuild/artifact/target/debs/*zip*/debs.zip`
  3. `sudo unzip debs.zip`
  
### Build Rest-API
  1. Modify `copy.sh` SDIR path to point to your downloaded debs from the previous step
  2. Execute ./build.sh
  3. The above should generate 2 Docker images, 1 which is used for local development on your VM and 1 which is used for deployment on a TOR
  4. Run `docker images` to check if rest-api dockers were generated <br/>
      		<pre>`REPOSITORY                     TAG                 IMAGE ID            CREATED             SIZE` <br/>
		      `rest-api-image                 latest              d2815fcb7356        2 days ago          222MB` <br/>
          `rest-api-image-test_local      latest              e62219a0bae2        2 days ago          222MB`</pre>
  5. `rest-api-image-test_local` is for local testing on a dev VM and `rest-api-image` is for TOR testing/deployment
  6. The production image is also stored into a compressed archive `rest-api-image.gz`
### Running Rest-API container
#### Run Rest-API container locally on a VM and execute unit tests
  1. `docker run -d --rm -p8090:8090 -p6379:6379 --name rest-api --cap-add NET_ADMIN --privileged -t rest-api-image-test_local:latest`
  2. `cd test`
  3. `python apitest.py`
  
####  Login to Rest-API container and check logs
  1. `docker exec -it rest-api bash`
  2. `vim /tmp/rest-api.err.log`
  
#### Run Rest-API container on a switch
  1. scp/copy over the generated archive(`rest-api-image.gz`) to your switch
  2. `docker load < rest-api-image.gz`
  3. `docker run -d -p=8090:8090/tcp -v /var/run/redis/redis.sock:/var/run/redis/redis.sock --name rest-api --cap-add NET_ADMIN --privileged -t rest-api-image:latest`
