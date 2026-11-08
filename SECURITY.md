# Security Best Practices for Adrenochain

This document outlines security best practices and hardening recommendations for deploying Adrenochain.

## 1. Configuration Security

### Environment Variables

**Never commit secrets to version control.** All sensitive configuration values (API keys, private keys, database passwords) should be stored in environment variables.

Use the `configs/env.example` template as a starting point:

```bash
# Generate secure random values for production
cp configs/env.example .env
# Edit .env with your actual secrets (via secure vault/secrets manager)
export $(cat .env | xargs)
```

### File Permissions

Critical system files should have restrictive permissions:

- **Private keys**: `600` (owner read/write only)
- **Configuration files**: `644` or `755` as appropriate
- **Binary executables**: Ensure proper ownership and execute permissions only for authorized users

Example from storage configuration:

```go
// Set restrictive file permissions
config.Storage.FilePermissions = 0600  // Owner read/write, no group/other access
```

### Default File Permissions

The minimum recommended default permissions are `755` for directories and `644` for files. Always apply the principle of least privilege in production environments.

## 2. Network Security

### Bootstrap Peer Configuration

When configuring bootstrap peers:

- **Use only known, trusted peer addresses**
- **Verify IP addresses against official sources**
- **Consider using DNS-based service discovery instead of hardcoded IPs**
- **Rate limit connections to prevent DDoS amplification**

Example secure bootstrap configuration:

```yaml
network:
  bootstrap_peers:
    - "dns4://mainnet.bootstrap.peer.com"
    # Use only verified, official peer addresses
```

### Network Topology

- Enable MDNS for local discovery with appropriate firewall rules
- Configure connection timeouts to prevent resource exhaustion
- Limit maximum peers to prevent resource exhaustion attacks

## 3. Storage Security

### Data Directory Isolation

The `Storage.DataDir` should:

- Be located in an isolated, non-world-writable path
- Have strict ownership (matching system user/process)
- Use filesystem features like SELinux/AppArmor labels when available

Example production configuration:

```yaml
storage:
  data_dir: "/var/lib/adrenochain/data"
  permissions: 0750  # Directory with owner rwx, group/other rx
```

## 4. Runtime Security

### Privilege Separation

Run blockchain nodes and critical services under **dedicated system accounts** with minimal permissions. Never run as root.

### Memory Protection

The codebase implements:

- Secure memory initialization for cryptographic operations
- Prevention of sensitive data leakage in logs
- Proper handling of private keys (never to stdout/stderr)

## 5. Development Security

### Code Review Checklist

Before committing code:

1. Check for hardcoded secrets or API keys
2. Verify no sensitive data is logged
3. Ensure proper error handling that doesn't leak internal information
4. Audit cryptographic implementations and key management
5. Review access control logic and authorization checks

## 6. Incident Response

### Security Event Logging

The system logs security events including:

- Authentication failures
- Unauthorized access attempts
- Configuration changes

Review log files regularly and set up alerting for suspicious patterns.

### Recommended Monitoring

- Monitor failed authentication attempts
- Track configuration changes
- Alert on unusual network traffic patterns
- Regular audit of file permissions and ownership

## 7. Third-Party Dependencies

Regularly update dependencies and review security advisories. Use the following practices:

- Enable security update notifications for Go modules
- Participate in the dependency vulnerability reporting process
- Review pull requests from external maintainers carefully

## Reporting Security Issues

If you discover a security vulnerability, please report it responsibly:

1. Do not create public issues that could aid exploitation
2. Contact the project maintainers through official communication channels
3. Include steps to reproduce (if applicable without exposing sensitive data)
4. Respect any responsible disclosure policies in effect

---

*This document is part of Adrenochain's security-conscious development philosophy. Always follow the principle of least privilege and defense in depth.*
