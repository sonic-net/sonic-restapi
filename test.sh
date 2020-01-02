
# Start rest api container
docker run -d --rm -p8080:8080 -p6379:6379 --name rest-api --cap-add NET_ADMIN --privileged -t rest-api-image:latest

# Wait until everything started
sleep 5

# Run out tests
cd test
python apitest.py

# Stop container
docker stop rest-api
