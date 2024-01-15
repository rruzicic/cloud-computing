# Cloud computing project

*Ratko Ružičić E2 102/2023*

# Usage

## Docker

- Run `docker compose up -d` in this directory.
- If you want to change the code you can specify `build: ./lib-[central | city]` instead of `image: rruzicic1/lib-[central | city]` in `docker-compose.yml`  

## Kubernetes

Start minikube

```sh
minikube start
```

Enable ingress controller 

```sh
minikube addons enable ingress
```

Setup infrastructure

```sh
kubectl apply -f k8s/
```

## Sending requests

Create user

```sh
curl --resolve "library.info:80:$( minikube ip )" -X POST -i "http://library.info/central/user?jmbg=1231231231232&name=jova&address=strazilovska
```

Get created user

```sh
curl --resolve "library.info:80:$( minikube ip )" -X GET -i "http://library.info/central/user?jmbg=1231231231232"
```

Lend a book

```sh
curl --resolve "library.info:80:$( minikube ip )" -X POST -i "http://library.info/bg/books/lending?jmbg=1231231231232&bookname=sidarta&isbn=I001SRB&author=hese"
```

Return a book

```sh
curl --resolve "library.info:80:$( minikube ip )" -X DELETE -i "http://library.info/bg/books/lending?jmbg=1231231231232&isbn=I001SRB"
```
