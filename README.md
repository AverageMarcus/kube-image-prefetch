# kube-image-prefetch

> Pre-pull all images, on all nodes, within a Kubernetes cluster

⚠️ Very early alpha release. Not yet for production use.

## Features

* Pull all images used by deployments in the cluster on all nodes
* Watch for new, changed or removed deployments and pre-fetch images on all nodes

## Install

```sh
kubectl apply -f https://raw.githubusercontent.com/AverageMarcus/kube-image-prefetch/master/manifest.yaml
```

## Building from source

With Docker:

```sh
make docker-build
```

Standalone:

```sh
make build
```

## Contributing

If you find a bug or have an idea for a new feature please [raise an issue](/issues/new) to discuss it.

Pull requests are welcomed but please try and follow similar code style as the rest of the project and ensure all tests and code checkers are passing.

Thank you 💙

## License

See [LICENSE](LICENSE)
