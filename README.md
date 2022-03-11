<p>
    <a href="https://kubeorbit.io/#gh-light-mode-only">
      <img src="https://raw.githubusercontent.com/teamcode-inc/public-resources/main/kubeOrbit_log.svg" width="400px">
    </a>
    <a href="https://kubeorbit.io/#gh-dark-mode-only">
      <img src="https://raw.githubusercontent.com/teamcode-inc/public-resources/main/KubeOrbit_for_Dark_Mode.svg" width="400px">
    </a>
</p>


Like KubeOrbit idea? ⭐ Give us a GitHub Star! ⭐

![work in progress badge](https://img.shields.io/badge/stability-work_in_progress-lightgrey.svg?style=flat-square)
[![Apache License 2.0](https://img.shields.io/github/license/teamcode-inc/kubeorbit?style=flat-square)](https://github.com/teamcode-inc/kubeorbit/blob/master/LICENSE)
[![Discord channel](https://img.shields.io/discord/930779108818956298?style=flat-square)](https://discord.gg/5XaTS9VArf)


**KubeOrbit** is an open-source tool that turns easy apps testing&debuging on Kubernetes in a new way. Our KubeOrbit is meant to create a channel automatically. You can *test* your *cloud-native* applications through this channel in a *hands-free* style.

It solves the following problems during integration tests:
- Under limited resource and restricted environment, developers in a team may be blocked by others who are testing their own functionalities, and it slows down the development progress.
- On the other hand, an unstable feature being deployed to a microservice may cause entire system crash.
<p align="center">
<img src="https://raw.githubusercontent.com/teamcode-inc/public-resources/main/kubeorbit_architecture.png" width="90%">
</p>


## Features
From now on, stop testing your application in local infra naively. Also, no more endless efforts in managing various cloud-based test environments.
- **KubeOrbit CLI**: just using one command, forward the traffic from in-cluster service to local service in a flash, no matter your service discovery is based on Eureka, Consul, Nacos or Kubernetes SVC.
- **Protocol support**: various protocols based on Layer-7 are supported. HTTP, gRPC, Thrift, Dubbo ...
- **Workload tag**: tag your workload by creating a new channel. Then your request can be routed to the right workload replica, where you can work with your mates to test&debug the same feature together.


## Getting Started
With the following tutorials:

**KubeOrbit CLI**:
* [Getting started](https://www.kubeorbit.io/docs/local-development)
* [How to build](https://www.kubeorbit.io/docs/how-to-build)


## Contributing
We're a warm and welcoming community of open source contributors. Please join. All types of contributions are welcome. Be sure to read our [Contributing Guide](./CONTRIBUTING.md) before submitting a Pull Request to the project.

## Community
#### Discord

Join the [KubeOrbit Discord channel](https://discord.gg/5XaTS9VArf) to chat with KubeOrbit developers and other users. This is a good place to learn about KubeOrbit, ask questions, and share your experiences.

## License
The KubeOrbit user space components are licensed under the [Apache License, Version 2.0](./LICENSE). 
