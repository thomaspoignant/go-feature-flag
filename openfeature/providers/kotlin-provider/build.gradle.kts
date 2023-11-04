// Top-level build file where you can add configuration options common to all sub-projects/modules.
plugins {
    id("com.android.application") version "8.1.2" apply false
    id("org.jetbrains.kotlin.android") version "1.9.0" apply false
    id("com.android.library") version "8.1.2" apply false
    id("org.jlleitschuh.gradle.ktlint") version "11.6.1" apply true
    id("io.github.gradle-nexus.publish-plugin") version "1.3.0" apply true
}

allprojects {
    extra["groupId"] = "org.gofeatureflag.openfeature"
    ext["version"] = "0.0.1"
}

group = project.extra["groupId"].toString()
version = project.extra["version"].toString()
nexusPublishing {
    this.repositories {
        sonatype {
            nexusUrl.set(uri("https://s01.oss.sonatype.org/service/local/"))
            snapshotRepositoryUrl.set(uri("https://s01.oss.sonatype.org/content/repositories/snapshots/"))
            username.set(System.getenv("OSSRH_USERNAME"))
            username.set(System.getenv("OSSRH_PASSWORD"))
        }
    }
}
