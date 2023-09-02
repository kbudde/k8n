# k8n

## Description
<img align="left" src="assets/k8n.png" alt="K8n gopher" width="90"/>

k8n (or "ken") is a Go-based project that combines the power of Kubernetes with the magic of a magician. It provides a universal operator that generates resources based on resources inside the Kubernetes cluster. The operator can be configured to watch resources based on API version, kind, and labels/annotations. On each change, the resources are provided as input to ytt overlays, to transform the resources into Kubernetes manifests you want to have. The manifests are then applied to the cluster with kapp.

## Examples
- Copying secrets from one namespace to others
- Create your own (simple) operator with the help of config maps or CRDs

See examples folder for quickstart

## Installation

K8N comes with prebuilt images and binaries. See github releases.

## Contributing
Contributions are welcome! If you'd like to contribute to k8n, please create an issue first and enable [lefthook](https://github.com/evilmartians/lefthook)
## License
k8n is released under the [MIT License](LICENSE).

## Acknowledgments
Special thanks to the Carvel team for their amazing work on [ytt](https://carvel.dev/ytt/) and [kapp](https://carvel.dev/kapp/), which are essential components of k8n.

## Contact
For any inquiries or feedback, please reach out via github issue or discussion
