FROM certbot/certbot AS build

# Install tools required to build the project
# We will need to run `docker build --no-cache .` to update those dependencies
WORKDIR app
COPY    certbot/entrypoint.py entrypoint.py
ENTRYPOINT ["python", "entrypoint.py"]
