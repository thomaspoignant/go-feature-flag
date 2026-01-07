---
title: '🔄 Runtime API Key Updates: Zero-Downtime Key Rotation in GO Feature Flag'
description: "Discover how GO Feature Flag's new runtime API key update feature enables you to rotate and update API keys without restarting the relay proxy, ensuring continuous availability and improved security."
authors: [thomaspoignant]
tags:
  [
    feature flags,
    api keys,
    security,
    runtime updates,
    zero downtime,
    key rotation,
  ]
image: https://gofeatureflag.org/assets/images/runtime-api-key-updates.jpg
---

# 🔄 Runtime API Key Updates: Zero-Downtime Key Rotation in GO Feature Flag

We're excited to announce a powerful new feature in GO Feature Flag that makes API key management easier and more secure: **Runtime API Key Updates**! 🎉

Managing API keys is a critical aspect of security, but it has traditionally required restarting services, causing potential downtime and service interruptions. With this new feature, you can now update, rotate, and manage API keys seamlessly while your relay proxy continues serving requests without interruption.

<!--truncate-->

## 🤔 The Challenge: API Key Management

API keys are essential for securing access to your feature flag relay proxy. However, managing them has always been challenging:

- **Key rotation** required service restarts, causing downtime
- **Adding new keys** meant taking the service offline
- **Removing compromised keys** couldn't be done immediately
- **No flexibility** to respond quickly to security incidents

These limitations made it difficult to follow security best practices like regular key rotation and immediate response to security threats.

## ✨ Introducing Runtime API Key Updates

GO Feature Flag now supports **updating API keys at runtime** without requiring a restart of the relay proxy. This feature works in both **default mode** and **flagset mode**, giving you the flexibility to manage keys dynamically.

### Key Benefits

- ✅ **Zero Downtime**: Update keys without interrupting service
- ✅ **Immediate Response**: Remove compromised keys instantly
- ✅ **Easy Rotation**: Rotate keys on a schedule without service impact
- ✅ **Flexible Management**: Add, remove, or update keys as needed
- ✅ **Automatic Detection**: Changes are detected and applied automatically

## 🚀 How It Works

The relay proxy continuously monitors your configuration file for changes. When it detects updates to API keys, it:

1. **Validates** the new configuration
2. **Updates** the internal API key mappings
3. **Applies** changes immediately
4. **Continues** serving requests without interruption

All of this happens automatically in the background, with no manual intervention required.

## 📝 Usage Examples

### Default Mode

In default mode, you can update `authorizedKeys` at runtime:

```yaml title="goff-proxy.yaml"
# Initial configuration
authorizedKeys:
  evaluation:
    - "key-1"
    - "key-2"
  admin:
    - "admin-key-1"

# Updated configuration (runtime update - no restart needed!)
authorizedKeys:
  evaluation:
    - "key-1"
    - "key-2"
    - "key-3"  # ✅ New key added
    - "key-4"  # ✅ Another new key
  admin:
    - "admin-key-1"
    - "admin-key-2"  # ✅ New admin key added
```

Simply update your configuration file, and the relay proxy will automatically apply the changes!

### Flagset Mode

In flagset mode, you can update API keys for each flag set individually:

```yaml title="goff-proxy.yaml"
flagSets:
  - name: team-a  # ✅ Name is required for runtime updates
    apiKeys:
      - "team-a-key-1"
      - "team-a-key-2"  # ✅ Can add new keys at runtime
    retrievers:
      - kind: file
        path: /flags/team-a-flags.yaml

  - name: team-b
    apiKeys:
      - "team-b-key-1"
      # ✅ Can remove keys at runtime by removing them from the config
    retrievers:
      - kind: s3
        bucket: team-b-flags
```

:::warning
**Important**: For flagset mode, your flag sets **must have a `name` configured** for runtime updates to work.  
Without a name, a UUID is generated automatically and runtime updates won't be possible.
:::

## 🔒 Security Best Practices

This feature enables you to follow security best practices more easily:

### Regular Key Rotation

Rotate your API keys on a schedule without service disruption:

```yaml
# Week 1
authorizedKeys:
  evaluation:
    - "key-old"
    - "key-new"

# Week 2 (after all clients migrated)
authorizedKeys:
  evaluation:
    - "key-new"  # ✅ Old key removed, only new key remains
```

### Immediate Response to Security Incidents

If a key is compromised, remove it immediately:

```yaml
# Before incident
authorizedKeys:
  evaluation:
    - "key-1"
    - "key-2"  # ⚠️ Compromised!

# After incident (immediate update)
authorizedKeys:
  evaluation:
    - "key-1"  # ✅ Compromised key removed instantly
```

### Gradual Key Migration

Add new keys alongside old ones, migrate clients gradually, then remove old keys:

```yaml
# Phase 1: Add new key
authorizedKeys:
  evaluation:
    - "key-old"
    - "key-new"  # ✅ New key added

# Phase 2: Migrate clients to new key
# (clients update their configuration)

# Phase 3: Remove old key
authorizedKeys:
  evaluation:
    - "key-new"  # ✅ Old key removed after migration
```

## ⚙️ Configuration Requirements

### Default Mode

In default mode, **only API keys can be updated at runtime**:
- ✅ `authorizedKeys.evaluation`
- ✅ `authorizedKeys.admin`
- ❌ All other configuration changes are ignored

### Flagset Mode

In flagset mode:
- ✅ API keys for each flag set can be updated
- ✅ Flag sets must have a `name` configured
- ❌ Other flagset configuration changes are not supported

## 🎯 Real-World Use Cases

### Use Case 1: Scheduled Key Rotation

**Scenario**: Your security policy requires rotating API keys every 90 days.

**Solution**: Update the configuration file with new keys, keep old keys temporarily, migrate clients, then remove old keys—all without downtime.

### Use Case 2: Team Onboarding

**Scenario**: A new team needs access to a flag set.

**Solution**: Simply add their API key to the flag set configuration. No restart needed, and they can start using the service immediately.

### Use Case 3: Security Incident Response

**Scenario**: An API key is suspected to be compromised.

**Solution**: Immediately remove the key from the configuration file. The relay proxy will stop accepting requests with that key within seconds, without any service interruption.

### Use Case 4: Multi-Tenant Key Management

**Scenario**: Managing API keys for multiple customers in flagset mode.

**Solution**: Update keys for individual flag sets independently. Each customer's keys can be managed separately without affecting others.

## 📚 Learn More

- 📖 [Runtime Configuration Updates Documentation](https://docs.gofeatureflag.org/relay-proxy/configure-relay-proxy#runtime-configuration-updates)
- 📖 [Flag Set Concepts](https://docs.gofeatureflag.org/concepts/flagset)
- 📖 [Configuration Guide](https://docs.gofeatureflag.org/relay-proxy/configure-relay-proxy)

## 🎉 Get Started Today

Runtime API key updates are available now in GO Feature Flag! This feature makes API key management more secure, flexible, and convenient.

Whether you're rotating keys, responding to security incidents, or managing access for multiple teams, runtime updates ensure your feature flag service remains available and secure.

### Resources

- 💬 [Community Discussions](https://gofeatureflag.org/slack)
- 🐛 [Report Issues](https://github.com/thomaspoignant/go-feature-flag/issues)
- 📚 [Full Documentation](https://docs.gofeatureflag.org)

We can't wait to see how this feature improves your API key management workflow! 🚀

