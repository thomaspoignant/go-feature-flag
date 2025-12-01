import org.jetbrains.kotlin.cli.jvm.main

plugins {
    kotlin("jvm") version "2.2.21"
    application
    id("com.github.johnrengelman.shadow") version "8.1.1"
    id("java")
}

group = "org.gofeatureflag.org"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
}

dependencies {
    testImplementation(kotlin("test"))
    implementation("dev.openfeature.contrib.providers:go-feature-flag:1.1.0")
    implementation("dev.openfeature:sdk:1.19.0")
}

tasks.test {
    useJUnitPlatform()
}

tasks.jar {
    manifest {
        attributes["Main-Class"] = "org.gofeatureflag.provider.server.example.MainKt"
    }
}

kotlin {
    jvmToolchain(11)
}

application {
    mainClass.set("org.gofeatureflag.provider.server.example.MainKt")
}