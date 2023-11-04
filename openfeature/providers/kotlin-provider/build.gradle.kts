// Top-level build file where you can add configuration options common to all sub-projects/modules.
plugins {
    id("com.android.application") version "8.1.2" apply false
    id("org.jetbrains.kotlin.android") version "1.9.0" apply false
    id("com.android.library") version "8.1.2" apply false
}

allprojects {
    extra["groupId"] = "org.gofeatureflag.openfeature"
    ext["version"] = "0.0.1"
}

group = project.extra["groupId"].toString()
version = project.extra["version"].toString()