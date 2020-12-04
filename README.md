# Config Model Support

This project provides a library and tools for managing YANG models in Go. It defines
the primary interfaces for YANG-based `ConfigModel`s, provides a `Repository` abstraction
for managing models, and includes a `PluginCompiler` which supports compiling and loading
models from YANG modules at runtime.

## Agent

The config agent is a tool that supports compiling and managing config models for
Kubernetes services. The agent is provided as a Docker image which can be deployed 
either as an init container or as a sidecar container. 

When deployed as an init container, the `config-agent plugin compile` command can be
used to compile model plugins for the primary service:

```bash
> go run github.com/onosproject/config-models/cmd/config-agent \
    plugin compile \--name test \
    --version 1.0.0 \
    --module test@2020-11-18=/root/plugins/test/test@2020-11-18.yang \
    --build-path /root/build/test
    --output-path /root/plugins
```

The agent can also be run as a sidecar to compile and manage config models throughout
the lifetime of a service. To run the agent:

```bash
> make serve
```

The agent implements a gRPC API exposing the config model repository to clients. The
`config-agent` CLI can be used to interact with the agent server:

To push a new model to the model repo, use the `repo push` sub-command:

```bash
> go run github.com/onosproject/config-models/cmd/config-agent \
    repo push \
    --name foo \
    --version 1.0.0 \
    --module test@2020-11-18=plugins/test/test@2020-11-18.yang
```

The agent will compile the model and add it to the repository. You can list the contents of the
repository with the `repo list` command or get information about a specific model with the
`repo get` command:

```bash
> go run github.com/onosproject/config-models/cmd/config-agent repo list
{
  "name": "foo",
  "version": "1.0.0",
  "modules": [
    {
      "name": "test",
      "organization": "",
      "version": "2020-11-18",
      "data": "bW9kdWxlIHRlc3QgewogIG5hbWVzcGFjZSAiaHR0cDovL29wZW5uZXR3b3JraW5nLm9yZy9vcmFuL3Rlc3QiOwogIHByZWZpeCB0MTsKCiAgb3JnYW5pemF0aW9uCiAgICAiT3BlbiBOZXR3b3JraW5nIEZvdW5kYXRpb24uIjsKICBjb250YWN0CiAgICAiQWRpYiBSYXN0ZWdhcm5pYSI7CiAgZGVzY3JpcHRpb24KICAgICJUbyBnZW5lcmF0ZSBKU09OIGZyb20gdGhpcyB1c2UgY29tbWFuZAogICAgIHB5YW5nIC1mIGp0b3h4IHRlc3QxLnlhbmcgfCBweXRob24zIC1tIGpzb24udG9vbCA+IHRlc3QxLmpzb24KICAgICBDb3BpZWQgZnJvbSBZYW5nVUlDb21wb25lbnRzIHByb2plY3QiOwoKICByZXZpc2lvbiAyMDIwLTExLTE4IHsKICAgIGRlc2NyaXB0aW9uCiAgICAgICJFeHRlbmRlZCB3aXRoIG5ldyBhdHRyaWJ1dGVzIG9uIGxlYWYyZCwgbGlzdDJiIjsKICAgIHJlZmVyZW5jZQogICAgICAiUkZDIDYwODciOwogIH0KCiAgY29udGFpbmVyIGNvbnQxYSB7CiAgICBkZXNjcmlwdGlvbgogICAgICAiVGhlIHRvcCBsZXZlbCBjb250YWluZXIiOwogICAgbGVhZiBsZWFmMWEgewogICAgICB0eXBlIHN0cmluZyB7CiAgICAgICAgbGVuZ3RoICIxLi44MCI7CiAgICAgIH0KICAgICAgZGVzY3JpcHRpb24KICAgICAgICAiZGlzcGxheSBuYW1lIHRvIHVzZSBpbiBHVUkgb3IgQ0xJIjsKICAgIH0KICAgIGxlYWYgbGVhZjJhIHsKICAgICAgdHlwZSBzdHJpbmcgewogICAgICAgIGxlbmd0aCAiMS4uMjU1IjsKICAgICAgfQogICAgICBkZXNjcmlwdGlvbgogICAgICAgICJ1c2VyIHBsYW5lIG5hbWUiOwogICAgfQogIH0KfQ=="
    }
  ],
  "plugin": {
    "name": "foo",
    "version": "1.0.0",
    "file": ""
  }
}
```

The JSON output above is the config model definition used to track the model within the model repo.
The model plugin can be loaded from within the agent container or any other container that shared
the model volume with the config agent. To load a model, simply call the `Load` function:

```go
import "github.com/onosproject/config-models/pkg/model"

...

myModel, err := model.Load("my-model-1.0.0.so")
```

The model object that's returned will be a generated implementation of the `ConfigModel` interface.
