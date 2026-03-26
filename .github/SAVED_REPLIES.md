# Saved Replies

These are standardized responses for the PentAGI Development Team to use when responding to Issues and Pull Requests. Using these templates helps maintain consistency in our communications and saves time.

Since GitHub currently does not support repository-wide saved replies, team members should maintain these individually. All responses are versioned for easier updates.

While these are templates, please customize them to fit the specific context and:
- Welcome new contributors
- Thank them for their contribution
- Provide context for your response
- Outline next steps

You can add these saved replies to [your personal GitHub account here](https://github.com/settings/replies).

## Issue Responses

### Issue: Already Fixed (v1)
```
Thank you for reporting this issue! This has been resolved in a recent release. Please update to the latest version (see our [releases page](https://github.com/vxcontrol/pentagi/releases)) and verify if the issue persists.

If you continue experiencing problems after updating, please:
1. Check your configuration against our documentation
2. Provide logs from both PentAGI and monitoring systems (Grafana/Langfuse)
3. Include details about your environment and enabled features
```

### Issue: Need More Information (v1)
```
Thank you for your report! To help us better understand and address your issue, please provide additional information:

1. PentAGI version and deployment method (Docker Compose/Custom)
2. Relevant logs from:
   - Docker containers
   - Grafana dashboards (if enabled)
   - Langfuse traces (if enabled)
3. Steps to reproduce the issue
4. Expected vs actual behavior

Please update your issue using our bug report template for consistency.
```

### Issue: Cannot Reproduce (v1)
```
Thank you for reporting this issue! Unfortunately, I cannot reproduce the problem with the provided information. To help us investigate:

1. Verify you're using the latest version
2. Provide your complete environment configuration
3. Share relevant logs and monitoring data
4. Include step-by-step reproduction instructions
5. Specify which AI agents were involved (Researcher/Developer/Executor)

Please update your issue with these details so we can better assist you.
```

### Issue: Expected Behavior (v1)
```
Thank you for your report! This appears to be the expected behavior because:

[Explanation of why this is working as designed]

If you believe this behavior should be different, please:
1. Describe your use case in detail
2. Explain why the current behavior doesn't meet your needs
3. Suggest alternative behavior that would work better

We're always open to improving PentAGI's functionality.
```

### Issue: Missing Template (v1)
```
Thank you for reporting this! To help us process your issue efficiently, please use our issue templates:

- [Bug Report Template](https://github.com/vxcontrol/pentagi/blob/master/.github/ISSUE_TEMPLATE/1-bug-report.md) for problems
- [Enhancement Template](https://github.com/vxcontrol/pentagi/blob/master/.github/ISSUE_TEMPLATE/2-enhancement.md) for suggestions

Please edit your issue to include the template information. This helps ensure we have all necessary details to assist you.
```

### Issue: PR Welcome (v1)
```
Thank you for raising this issue! We welcome contributions from the community.

If you'd like to implement this yourself:
1. Check our [contribution guidelines](CONTRIBUTING.md)
2. Review the architecture documentation
3. Consider security implications (especially for AI agent modifications)
4. Include tests and documentation
5. Update monitoring/analytics as needed

Feel free to ask questions if you need guidance. We're here to help!
```

## PR Responses

### PR: Ready to Merge (v1)
```
Excellent work! This PR meets our quality standards and I'll proceed with merging it.

If you're interested in further contributions, check our:
- [Help Wanted Issues](https://github.com/vxcontrol/pentagi/labels/help-wanted)
- [Good First Issues](https://github.com/vxcontrol/pentagi/labels/good-first-issue)

Thank you for improving PentAGI!
```

### PR: Needs Work (v1)
```
Thank you for your contribution! A few items need attention before we can merge:

[List specific items that need addressing]

Common requirements:
- Tests for new functionality
- Documentation updates
- Security considerations
- Performance impact assessment
- Monitoring/analytics integration

Please update your PR addressing these points. Let us know if you need any clarification.
```

### PR: Missing Template (v1)
```
Thank you for your contribution! Please update your PR to use our [PR template](https://github.com/vxcontrol/pentagi/blob/master/.github/PULL_REQUEST_TEMPLATE.md).

The template helps ensure we have:
- Clear description of changes
- Testing information
- Security considerations
- Documentation updates
- Deployment notes

This helps us review your changes effectively.
```

### PR: Missing Issue (v1)
```
Thank you for your contribution! We require an associated issue for each PR to:
- Discuss approach before implementation
- Track related changes
- Maintain clear project history

Please:
1. [Create an issue](https://github.com/vxcontrol/pentagi/issues/new/choose)
2. Link it to this PR
3. Update the PR description with the issue reference

This helps us maintain good project organization.
```

### PR: Inactive (v1)
```
This PR has been inactive for a while. To keep our review process efficient:

1. If you're still working on this:
   - Let us know your timeline
   - Update with latest main branch
   - Address any existing feedback

2. If you're no longer working on this:
   - We can close it
   - Someone else can pick it up

Please let us know your preference within the next week.
```

### General: Need Help (v1)
```
I need additional expertise on this. Pinging:
- @asdek for technical review
- @security-team for security implications
- @ai-team for AI agent behavior
- @infra-team for infrastructure changes

[Specific questions or concerns that need addressing]
```
