# OpenVAS via a Custom Pentest Image

> [!WARNING]
> This is an advanced custom-image workflow, not built-in OpenVAS support in PentAGI.
> PentAGI does not install, configure, feed-sync, or orchestrate OpenVAS/GVM for you.

Issue [#69](https://github.com/vxcontrol/pentagi/issues/69) asks about OpenVAS support. The current maintainer-recommended workaround is to build your own pentest image on top of `vxcontrol/kali-linux:systemd`, install OpenVAS/GVM there, point PentAGI at that image with `DOCKER_DEFAULT_IMAGE_FOR_PENTEST`, and update prompts so agents know when and how to use it.

The default PentAGI pentest image remains `vxcontrol/kali-linux`. This guide documents how to replace it with your own OpenVAS-ready image without implying first-class native OpenVAS support.

## When to Use This Workflow

Use this approach if you want to experiment with OpenVAS/GVM inside PentAGI's pentest containers while keeping the change local to your own deployment.

Do not use this guide as evidence that PentAGI has native OpenVAS orchestration, a built-in installer, or a maintained OpenVAS API integration.

## Build a Custom Pentest Image

OpenVAS/GVM expects systemd out of the box, so the maintainer recommendation is to start from `vxcontrol/kali-linux:systemd` rather than the default `vxcontrol/kali-linux` image.

### Option 1: Install Packages in a Custom Image

This is the shortest path when the needed packages are available in your chosen base image state.

```dockerfile
FROM vxcontrol/kali-linux:systemd

RUN apt update && apt install -y openvas gvm \
    && rm -rf /var/lib/apt/lists/*

# Provide your own modified copy of the upstream entrypoint that starts both
# systemd and the GVM/OpenVAS services you need.
COPY container-entrypoint.sh /usr/local/bin/container-entrypoint
RUN chmod +x /usr/local/bin/container-entrypoint
```

> [!NOTE]
> Treat this Dockerfile as an illustrative starting point. Package names, repositories, and service startup details can vary by distro state. If package installation fails, use the source-build route below instead of assuming PentAGI is missing a feature.

### Option 2: Build from Source

If the package-install path does not work for your image state, follow Greenbone's source-build example instead:

- [Greenbone OpenVAS scanner Debian example Dockerfile](https://github.com/greenbone/openvas-scanner/blob/main/.docker/railguards/debian_stable.Dockerfile)

Use the Greenbone build steps as a reference for compiling the scanner, but adapt them into an image that still starts from `vxcontrol/kali-linux:systemd` and uses a working entrypoint for systemd plus the GVM/OpenVAS services you rely on.

### Entrypoint Caveat

The upstream base image already uses a custom entrypoint at `/usr/local/bin/container-entrypoint`.

- Base image repository: [vxcontrol/kali-linux-image](https://github.com/vxcontrol/kali-linux-image)
- Upstream entrypoint script: [container-entrypoint.sh](https://github.com/vxcontrol/kali-linux-image/blob/master/container-entrypoint.sh)

Create and maintain your own modified copy of that entrypoint during your image build so it starts both systemd and the GVM/OpenVAS services you need. PentAGI does not ship a ready-made `container-entrypoint.sh` for this workflow.

## Connect the Custom Image to PentAGI

1. Build and tag your image:

   ```bash
   docker build -t myorg/kali-linux:openvas .
   ```

2. Set the PentAGI pentest image in `.env`:

   ```bash
   DOCKER_DEFAULT_IMAGE_FOR_PENTEST=myorg/kali-linux:openvas
   ```

3. Restart PentAGI so new pentest tasks use the updated default image.

4. Start a new pentest task or flow after the restart.

> [!NOTE]
> If a user explicitly requests a different Docker image in a task, PentAGI may use that image instead. `DOCKER_DEFAULT_IMAGE_FOR_PENTEST` affects automatic image selection, not explicit per-task overrides.

### Worker-Node Deployments

If you run a distributed setup, use the same custom image as the pentest image on the worker-node side as well. The worker-node guide shows where PentAGI expects the pentest image to be configured: [Worker Node Setup](worker_node.md).

## Prompt Guidance

After the image is available, tell the agents that OpenVAS/GVM exists in your environment.

- Add reusable guidance in `Settings -> Prompts`.
- If you already use custom flow prompts, add the same expectations there so flow-specific behavior stays consistent.

Small example prompt text:

```text
OpenVAS/GVM is available in this deployment's pentest image.
Before using it, verify the required services and scanner are ready inside the container.
Use OpenVAS when broad vulnerability scanning is appropriate, and save scan outputs and exported artifacts under /work.
If OpenVAS is unavailable or not ready, continue with other tools and report the limitation clearly.
```

## Limitations and Expectations

- This is a self-managed custom image workflow, not built-in OpenVAS support.
- PentAGI does not install OpenVAS, initialize feeds, manage service readiness, or expose a dedicated OpenVAS API layer in this setup.
- Prompt changes are still required so agents know when OpenVAS is available and when it is appropriate to use it.
- New tasks created after the restart pick up the updated default image more reliably than already-running work.

## Troubleshooting

### `apt install openvas gvm` Fails or the Packages Are Missing

Package names and repository availability can change across distro states. Verify your package sources, then fall back to the source-build route if the package-install path is unavailable.

### PentAGI Still Uses the Old Pentest Image

Confirm `DOCKER_DEFAULT_IMAGE_FOR_PENTEST` is set correctly, restart PentAGI, and start a new pentest task or flow. Existing work may continue using containers created from the previous image.

### Agents Ignore OpenVAS

Refine your prompt wording, make the OpenVAS/GVM availability explicit, and review execution traces to see whether the agent recognized the tool but chose a different path.

### GVM or System Services Do Not Start

Re-check your custom `/usr/local/bin/container-entrypoint` logic. The workflow depends on your image starting both systemd and the OpenVAS/GVM services you need before the agent tries to use them.
