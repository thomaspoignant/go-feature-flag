# Seamlessly Update API Keys at Runtime in GO Feature Flag (No Restarts Needed)

January 7, 2026 ·

<!-- -->

4 min read

[![Thomas Poignant](https://github.com/thomaspoignant.png)](https://github.com/thomaspoignant)

[Thomas Poignant](https://github.com/thomaspoignant)

Creator of GO Feature Flag

![In-Process OpenFeature Providers Banner](/assets/images/banner-e0b5754b6ec70d57c3f1d562b7121763.jpg)

We're excited to announce a new feature in GO Feature Flag that makes API key management easier and more secure: **Runtime API Key Updates**! 🎉

We've heard from the community that having to restart the relay proxy for every API key update or rotation was a real pain point. Managing API keys is crucial for security, but needing a restart for each change led to unnecessary downtime and interruptions.<br /><!-- -->With this new feature, you can now update, rotate, and manage API keys on the fly—your relay proxy keeps running, and requests are served continuously, with no need for restarts.

## 🤔 The Challenge: API Key Management[​](#-the-challenge-api-key-management "Direct link to 🤔 The Challenge: API Key Management")

API keys are essential for securing access to your GO Feature Flag relay proxy. However, managing them has always been challenging:

* **Key rotation** required service restarts, causing downtime
* **Adding new keys** meant taking the service offline
* **Removing compromised keys** couldn't be done immediately
* **No flexibility** to respond quickly to security incidents

These limitations made it difficult to follow security best practices like regular key rotation and immediate response to security threats.

## ✨ Introducing Runtime API Key Updates[​](#-introducing-runtime-api-key-updates "Direct link to ✨ Introducing Runtime API Key Updates")

GO Feature Flag starting `v1.50.0` supports **updating API keys at runtime** without requiring a restart of the relay proxy.<br /><!-- -->This feature works in both **default mode** and **flagset mode**, giving you the flexibility to manage keys dynamically.

info

This feature is available starting from GO Feature Flag `v1.50.0` and above.

### Key Benefits[​](#key-benefits "Direct link to Key Benefits")

* ✅ **Zero Downtime**: Update keys without interrupting service
* ✅ **Immediate Response**: Remove compromised keys instantly
* ✅ **Easy Rotation**: Rotate keys as soon as you change your configuration file
* ✅ **Flexible Management**: Add, remove, or update keys as needed
* ✅ **Automatic Detection**: Changes are detected and applied automatically

## 🚀 How It Works[​](#-how-it-works "Direct link to 🚀 How It Works")

The relay proxy continuously monitors your configuration file for changes. When it detects updates to API keys, it:

1. **Validates** the new configuration *(if invalid your new configuration will be ignored)*
2. **Updates** the internal API key mappings
3. **Applies** changes immediately
4. **Continues** serving requests without interruption

All of this happens automatically in the background, with no manual intervention required.

## ⚙️ Configuration Requirements[​](#️-configuration-requirements "Direct link to ⚙️ Configuration Requirements")

### Default Mode[​](#default-mode "Direct link to Default Mode")

In default mode, **only API keys can be updated at runtime**:

* ✅ `authorizedKeys.evaluation`
* ✅ `authorizedKeys.admin`
* ❌ All other configuration changes are ignored

#### 📝 Usage Example[​](#-usage-example "Direct link to 📝 Usage Example")

goff-proxy.yaml

```
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

### Flagset Mode[​](#flagset-mode "Direct link to Flagset Mode")

In flagset mode:

* ✅ API keys for each flag set can be updated
* ✅ Flag sets must have a `name` configured
* ❌ Other flagset configuration changes are not supported

#### 📝 Usage Example[​](#-usage-example-1 "Direct link to 📝 Usage Example")

goff-proxy.yaml

```
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

warning

**Important**: For flagset mode, your flag sets **must have a `name` configured** for runtime updates to work.<br /><!-- -->Without a name (or if you use `default` as the name), runtime updates won't be possible since we are not able to target which flagset has changed.

## 🎯 Real-World Use Cases[​](#-real-world-use-cases "Direct link to 🎯 Real-World Use Cases")

### Use Case 1: Multi-Tenant Key Management[​](#use-case-1-multi-tenant-key-management "Direct link to Use Case 1: Multi-Tenant Key Management")

**Scenario**: Managing API keys for multiple customers in flagset mode.

**Solution**: Update keys for individual flag sets independently. Each customer's keys can be managed separately without affecting others.

### Use Case 2: Scheduled Key Rotation[​](#use-case-2-scheduled-key-rotation "Direct link to Use Case 2: Scheduled Key Rotation")

**Scenario**: Your security policy requires rotating API keys every 90 days.

**Solution**: Update the configuration file with new keys, keep old keys temporarily, migrate clients, then remove old keys—all without downtime.

### Use Case 3: Team Onboarding[​](#use-case-3-team-onboarding "Direct link to Use Case 3: Team Onboarding")

**Scenario**: A new team needs access to a flag set.

**Solution**: Simply add their API key to the flag set configuration. No restart needed, and they can start using the service immediately.

## 📚 Learn More[​](#-learn-more "Direct link to 📚 Learn More")

* 📖 [Runtime Configuration Updates Documentation](https://docs.gofeatureflag.org/relay-proxy/configure-relay-proxy#runtime-configuration-updates)
* 📖 [Flag Set Concepts](https://docs.gofeatureflag.org/concepts/flagset)
* 📖 [Configuration Guide](https://docs.gofeatureflag.org/relay-proxy/configure-relay-proxy)

## 🎉 Get Started Today[​](#-get-started-today "Direct link to 🎉 Get Started Today")

Runtime API key updates are available in GO Feature Flag starting version `v1.50.0`! This feature makes API key management more secure, flexible, and convenient.

Whether you're rotating keys, responding to security incidents, or managing access for multiple teams, runtime updates ensure your feature flag service remains available and secure.

### Resources[​](#resources "Direct link to Resources")

* 💬 [Community Discussions](https://gofeatureflag.org/slack)
* 🐛 [Report Issues](https://github.com/thomaspoignant/go-feature-flag/issues)
* 📚 [Full Documentation](https://docs.gofeatureflag.org)

We can't wait to see how this feature improves your API key management workflow! 🚀

**Tags:**

* [api keys](https://gofeatureflag.org/blog/tags/api-keys)
* [runtime updates](https://gofeatureflag.org/blog/tags/runtime-updates)
* [key rotation](https://gofeatureflag.org/blog/tags/key-rotation)

[Edit this page](https://github.com/thomaspoignant/go-feature-flag/tree/main/website/blog/2026-01-07-runtime-api-key-updates/index.md)
