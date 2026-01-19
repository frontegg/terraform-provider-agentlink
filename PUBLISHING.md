# Publishing to HashiCorp Terraform Registry

This document describes the requirements and steps to publish the `agentlink-terraform-provider` to the [HashiCorp Terraform Registry](https://registry.terraform.io/).

## Prerequisites

### 1. GitHub Repository Setup

The repository must meet these requirements:

- **Repository Name**: Must follow the pattern `terraform-provider-{NAME}`
  - Current: `agentlink-terraform-provider`
  - **Required**: Rename to `terraform-provider-agentlink`

- **Public Repository**: The repository must be public for the Terraform Registry to access it

- **GitHub Releases**: Releases must be created with semantic versioning tags (e.g., `v0.1.0`)

### 2. GPG Key for Signing

The Terraform Registry requires all provider releases to be signed with GPG. You need to:

#### Generate a GPG Key (if you don't have one)

```bash
# Generate a new GPG key
gpg --full-generate-key

# Choose:
# - RSA and RSA (default)
# - 4096 bits
# - Key does not expire (or set expiration)
# - Your name and email

# List your keys to get the fingerprint
gpg --list-secret-keys --keyid-format=long

# Export the public key (needed for Terraform Registry)
gpg --armor --export YOUR_KEY_ID > public-key.asc

# Export the private key (needed for GitHub Actions)
gpg --armor --export-secret-keys YOUR_KEY_ID > private-key.asc
```

#### Add GPG Key to GitHub Secrets

Add these secrets to your GitHub repository (`Settings > Secrets and variables > Actions`):

| Secret Name | Description |
|-------------|-------------|
| `GPG_PRIVATE_KEY` | The entire contents of `private-key.asc` |
| `GPG_PASSPHRASE` | The passphrase for your GPG key |

### 3. Register with Terraform Registry

1. Go to [registry.terraform.io](https://registry.terraform.io/)
2. Sign in with your GitHub account
3. Click **Publish** > **Provider**
4. Select the GitHub organization/user: `frontegg`
5. Select the repository: `terraform-provider-agentlink`
6. **Upload your GPG public key** when prompted
7. Complete the registration

### 4. Required GitHub Repository Secrets

Configure these secrets in your GitHub repository:

| Secret | Required | Description |
|--------|----------|-------------|
| `GPG_PRIVATE_KEY` | Yes | GPG private key for signing releases |
| `GPG_PASSPHRASE` | Yes | Passphrase for the GPG key |
| `GITHUB_TOKEN` | Auto | Automatically provided by GitHub Actions |

## Repository Naming

**Important**: The HashiCorp Terraform Registry requires repositories to follow the naming convention:

```
terraform-provider-{NAME}
```

Your current repository is named `agentlink-terraform-provider`, which needs to be renamed to:

```
terraform-provider-agentlink
```

### How to Rename the Repository

1. Go to GitHub repository settings
2. Under "Repository name", change to `terraform-provider-agentlink`
3. Update any references in documentation and workflows

After renaming, users will install the provider as:

```hcl
terraform {
  required_providers {
    agentlink = {
      source  = "frontegg/agentlink"
      version = "~> 0.1"
    }
  }
}
```

## Release Process

Once configured, the release process is automated:

1. **Push to master/main** triggers the release workflow
2. **Version is incremented** automatically (minor version bump)
3. **Git tag is created** with the new version
4. **GoReleaser builds** binaries for all platforms
5. **Artifacts are signed** with your GPG key
6. **GitHub Release is created** with all artifacts
7. **Terraform Registry syncs** the new release automatically

### Manual Release (if needed)

```bash
# Create and push a tag manually
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# Or use GoReleaser locally
export GPG_FINGERPRINT="YOUR_GPG_FINGERPRINT"
goreleaser release --clean
```

## Terraform Registry Webhook

The Terraform Registry uses GitHub webhooks to detect new releases. Ensure:

1. The webhook is configured (done automatically during registration)
2. Releases are not marked as "draft" (already configured in `.goreleaser.yml`)
3. Releases have proper semantic version tags (e.g., `v1.0.0`, `v1.1.0`)

## Release Artifacts

Each release must include these artifacts (automatically created by GoReleaser):

| Artifact | Description |
|----------|-------------|
| `terraform-provider-agentlink_VERSION_OS_ARCH.zip` | Provider binary archives |
| `terraform-provider-agentlink_VERSION_SHA256SUMS` | Checksums file |
| `terraform-provider-agentlink_VERSION_SHA256SUMS.sig` | GPG signature of checksums |

## Supported Platforms

The provider is built for these platforms (configured in `.goreleaser.yml`):

| OS | Architectures |
|----|---------------|
| Linux | amd64, 386, arm, arm64 |
| macOS (Darwin) | amd64, arm64 |
| Windows | amd64, 386, arm, arm64 |
| FreeBSD | amd64, 386, arm, arm64 |

## Verification

After publishing, verify the provider appears on the registry:

```
https://registry.terraform.io/providers/frontegg/agentlink/latest
```

Test installation:

```bash
terraform init
```

## Troubleshooting

### Release Not Appearing on Registry

1. Check that the release is not marked as "draft"
2. Verify the tag follows semantic versioning (e.g., `v1.0.0`)
3. Ensure all required artifacts are present in the release
4. Check the webhook delivery in GitHub settings

### GPG Signature Errors

1. Verify the GPG public key is uploaded to Terraform Registry
2. Check that `GPG_PRIVATE_KEY` secret contains the full private key
3. Ensure `GPG_PASSPHRASE` is correct

### Build Failures

1. Check GoReleaser logs in GitHub Actions
2. Verify Go version compatibility
3. Ensure all dependencies are available

## Summary Checklist

Before publishing, ensure you have:

- [ ] Renamed repository to `terraform-provider-agentlink`
- [ ] Generated GPG key pair
- [ ] Added `GPG_PRIVATE_KEY` to GitHub secrets
- [ ] Added `GPG_PASSPHRASE` to GitHub secrets
- [ ] Uploaded GPG public key to Terraform Registry
- [ ] Registered provider with Terraform Registry
- [ ] Made repository public
- [ ] Tested release workflow locally or in a test branch

## Support

For issues with:
- **Terraform Registry**: [HashiCorp Support](https://support.hashicorp.com/)
- **This Provider**: [GitHub Issues](https://github.com/frontegg/terraform-provider-agentlink/issues)
