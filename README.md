# et-operator-extension
https://github.com/AliyunContainerService/et-operator

Some distributed deep learning training framework like horovod support elastic training, which enables training job scale up and down the number of workers dynamically at runtime without interrupting the training process.

Et-operator provides a set of Kubernetes Custom Resource Definition that makes it easy to run horovod or AIACC elastic training in kubernetes. After submit a training job, you can scaleIn and scaleOut workers during training on demand, which can make your training job more elasticity and efficient.

et-operator-extension is a kube-queue extension for et-operator.
