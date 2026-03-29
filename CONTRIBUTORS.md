# Contributors

This document recognizes all individuals who have contributed to PentAGI. Their work across 370+ commits over 18 months has shaped this project into what it is today.

## Core Team

### Project Lead & Backend Development
- [@asdek](https://github.com/asdek) (Dmitry Nagibin) - Architecture, backend infrastructure, agent system, provider integrations, observability, and project coordination

### Frontend Development
- [@sirozha](https://github.com/sirozha) (Sergey Kozyrenko) - React UI implementation, settings interfaces, GraphQL integration, and frontend architecture

### Backend Development
- [@zavgorodnii](https://github.com/zavgorodnii) (Andrei Zavgorodnii) - Graphiti integration, patch refiner, knowledge graph implementation

---

## External Contributors

We are deeply grateful to the following individuals for their contributions to PentAGI:

### Feature Contributors

#### [@mason5052](https://github.com/mason5052) (Mason Kim)
**Contributions:** Significant testing infrastructure improvements and bug fixes
- Added comprehensive unit test coverage across multiple packages ([PR#199](https://github.com/vxcontrol/pentagi/pull/199), [PR#198](https://github.com/vxcontrol/pentagi/pull/198), [PR#200](https://github.com/vxcontrol/pentagi/pull/200), [PR#201](https://github.com/vxcontrol/pentagi/pull/201), [PR#202](https://github.com/vxcontrol/pentagi/pull/202), [PR#214](https://github.com/vxcontrol/pentagi/pull/214), [PR#213](https://github.com/vxcontrol/pentagi/pull/213))
- Test coverage for: config, version, terminal, server response, embeddings, graph context, executor helpers, custom JSON types ([PR#170](https://github.com/vxcontrol/pentagi/pull/170)), context registry ([PR#171](https://github.com/vxcontrol/pentagi/pull/171)), executor terminal utilities ([PR#172](https://github.com/vxcontrol/pentagi/pull/172)), search tools ([PR#153](https://github.com/vxcontrol/pentagi/pull/153))
- Test coverage for LLM providers (DeepSeek, GLM, Kimi, Qwen) ([PR#189](https://github.com/vxcontrol/pentagi/pull/189))
- Fixed critical bugs:
  - Detached command context isolation ([PR#179](https://github.com/vxcontrol/pentagi/pull/179))
  - Agent chain iteration cap and repeating detector ([PR#178](https://github.com/vxcontrol/pentagi/pull/178), [PR#180](https://github.com/vxcontrol/pentagi/pull/180))
  - HTTP client global mutation in search tools ([PR#151](https://github.com/vxcontrol/pentagi/pull/151))
  - Browser graceful screenshot failure handling ([PR#150](https://github.com/vxcontrol/pentagi/pull/150))
  - Silent error handling in GetTool ([PR#152](https://github.com/vxcontrol/pentagi/pull/152))
  - OAuth callback missing return ([PR#127](https://github.com/vxcontrol/pentagi/pull/127))
  - OAuth state field validation ([PR#125](https://github.com/vxcontrol/pentagi/pull/125), [PR#120](https://github.com/vxcontrol/pentagi/pull/120))
  - Langfuse TLS configuration respect ([PR#132](https://github.com/vxcontrol/pentagi/pull/132))
  - CA private key cleanup in entrypoint ([PR#168](https://github.com/vxcontrol/pentagi/pull/168))
  - Google search options unused ([PR#167](https://github.com/vxcontrol/pentagi/pull/167))
  - Resource leaks and unbounded allocation in tools ([PR#141](https://github.com/vxcontrol/pentagi/pull/141))
  - Debug console.log removal ([PR#141](https://github.com/vxcontrol/pentagi/pull/141))
  - Swagger missing quote ([PR#140](https://github.com/vxcontrol/pentagi/pull/140))
  - Terminal typo correction ([PR#164](https://github.com/vxcontrol/pentagi/pull/164))
- Documentation improvements: fixed typos and grammar issues ([PR#121](https://github.com/vxcontrol/pentagi/pull/121))

#### [@niuqun2003](https://github.com/niuqun2003) (niuqun2003)
**Contributions:** Chinese LLM provider ecosystem integration ([PR#154](https://github.com/vxcontrol/pentagi/pull/154))
- Added support for DeepSeek, GLM, Kimi, and Qwen LLM providers
- Implemented provider configurations and API integrations

#### [@Priyanka-2725](https://github.com/Priyanka-2725) (Priyanka Singh)
**Contributions:** AWS Bedrock provider improvements, Sploitus inegration implementation and bug fixes
- Fixed Bedrock toolConfig runtime failures ([PR#166](https://github.com/vxcontrol/pentagi/pull/166))
- Implemented better exploit finding capabilities by Sploitus ([PR#133](https://github.com/vxcontrol/pentagi/pull/133))
- Fixed terminal command handling logic ([PR#124](https://github.com/vxcontrol/pentagi/pull/124))

#### [@Alex-wuhu](https://github.com/Alex-wuhu) (Alex)
**Contributions:** Novita AI integration ([PR#162](https://github.com/vxcontrol/pentagi/pull/162))
- Added Novita AI as optional LLM provider
- Updated Novita models configuration

#### [@efe-arv](https://github.com/efe-arv) (Efe Büken)
**Contributions:** HTTP client timeout configuration ([PR#205](https://github.com/vxcontrol/pentagi/pull/205))
- Added configurable timeout to HTTP client for improved reliability

#### [@manusjs](https://github.com/manusjs) (manusjs)
**Contributions:** Bedrock tool configuration fix ([PR#196](https://github.com/vxcontrol/pentagi/pull/196))
- Fixed Bedrock to always include toolConfig when messages contain toolUse/toolResult blocks

#### [@stoykovstoyk](https://github.com/stoykovstoyk) (Stoyko Stoykov)
**Contributions:** SearXNG meta search engine integration ([PR#53](https://github.com/vxcontrol/pentagi/pull/53))
- Added SearXNG as a search engine option
- Implemented tool integration for meta search capabilities

#### [@kaikreuzer](https://github.com/kaikreuzer) (Kai Kreuzer)
**Contributions:** AWS credentials support ([PR#90](https://github.com/vxcontrol/pentagi/pull/90))
- Added support for temporary AWS credentials
- Enhanced AWS session token handling

#### [@mrigankad](https://github.com/mrigankad) (Mriganka Dey)
**Contributions:** Security and bug fixes ([PR#104](https://github.com/vxcontrol/pentagi/pull/104))
- Various security improvements and bug fixes across the codebase

#### [@salmanmkc](https://github.com/salmanmkc) (Salman Chishti)
**Contributions:** GitHub Actions modernization
- Upgraded GitHub Actions to latest versions ([PR#112](https://github.com/vxcontrol/pentagi/pull/112))
- Upgraded GitHub Actions for Node 24 compatibility ([PR#111](https://github.com/vxcontrol/pentagi/pull/111))

### Bug Fixes & Improvements

#### [@Vaibhavee89](https://github.com/Vaibhavee89) (Vaibhavee Singh)
**Contributions:** Documentation enhancement
- Added external network access configuration guide to README

#### [@PeterDaveHello](https://github.com/PeterDaveHello) (Peter Dave Hello)
**Contributions:** Dockerfile optimization ([PR#50](https://github.com/vxcontrol/pentagi/pull/50))
- Removed unnecessary cleanup steps in Dockerfile for improved build efficiency

#### [@s-b-repo](https://github.com/s-b-repo) (S.B)
**Contributions:** Security improvements
- Implemented file size limit and path escaping for enhanced security
- Fixed typo in executor.go ('Incoming')

#### [@SkyFlyingMouse](https://github.com/SkyFlyingMouse) (SkyFlyingMouse)
**Contributions:** Code quality ([PR#128](https://github.com/vxcontrol/pentagi/pull/128))
- Fixed Docker client constant name typo in backend

#### [@haosenwang1018](https://github.com/haosenwang1018) (Sense_wang)
**Contributions:** Development environment ([PR#163](https://github.com/vxcontrol/pentagi/pull/163))
- Expanded .gitignore with IDE and OS patterns for cleaner repository

#### [@hhktony](https://github.com/hhktony) (Tony Xu)
**Contributions:** Documentation ([PR#32](https://github.com/vxcontrol/pentagi/pull/32))
- README.md improvements and clarifications

---

## Acknowledgments

This project exists because of the collective effort of everyone listed above. From major feature implementations to small bug fixes, every contribution has made PentAGI better.

---

## How to Contribute

Interested in contributing to PentAGI? We welcome contributions of all kinds:

- Bug reports and fixes
- New features and enhancements
- Documentation improvements
- Testing and QA
- Code reviews

Please see our [CONTRIBUTING.md](CONTRIBUTING.md) guide for more information.

---

**Note:** Due to repository history rewriting on March 29, 2026, to resolve licensing matters, individual commit history is no longer visible in GitHub's interface. This document preserves the record of all contributions made during the project's development from January 2025 to March 2026.
