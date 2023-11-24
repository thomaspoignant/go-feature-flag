plugins {
    id("com.android.library")
    id("org.jetbrains.kotlin.android")
    id("signing")
    id("maven-publish")
    id("org.jlleitschuh.gradle.ktlint")
}

val releaseVersion = project.extra["version"].toString()

android {
    namespace = "org.gofeatureflag.openfeature"
    compileSdk = 33

    defaultConfig {
        minSdk = 21
        version = releaseVersion
        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
    }

    buildTypes {
        release {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }
    kotlinOptions {
        jvmTarget = JavaVersion.VERSION_11.toString()
    }
    publishing {
        singleVariant("release") {
            withJavadocJar()
            withSourcesJar()
        }
    }
}
publishing {
    publications {
        register<MavenPublication>("release") {
            groupId = project.extra["groupId"].toString()
            artifactId = "gofeatureflag-kotlin-provider"
            version = releaseVersion

            pom {
                name.set("GO Feature Flag OpenFeature Provider for Android")
                description.set(
                    "This is the Android provider implementation of OpenFeature for GO Feature Flag."
                )
                url.set("https://gofeatureflag.org")
                licenses {
                    license {
                        name.set("The Apache License, Version 2.0")
                        url.set("http://www.apache.org/licenses/LICENSE-2.0.txt")
                    }
                }
                developers {
                    developer {
                        id.set("thomaspoignant")
                        name.set("Thomas Poignant")
                        email.set("thomas.poignant@gofeatureflag.org")
                    }
                }
                scm {
                    connection.set(
                        "scm:git:https://github.com/thomaspoignant/go-feature-flag.git"
                    )
                    developerConnection.set(
                        "scm:git:ssh://git@github.com:thomaspoignant/go-feature-flag.git"
                    )
                    url.set("https://github.com/thomaspoignant/go-feature-flag/tree/main/openfeature/providers/kotlin-provider")
                }
            }

            afterEvaluate {
                from(components["release"])
            }
        }
    }
}

dependencies {
    implementation("dev.openfeature:kotlin-sdk:0.1.0")
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    implementation("com.google.code.gson:gson:2.10.1")
    implementation("dev.gustavoavila:java-android-websocket-client:2.0.2")
    testImplementation("junit:junit:4.13.2")
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:1.7.3")
    testImplementation("com.squareup.okhttp3:mockwebserver:4.12.0")
}

signing {
    sign(publishing.publications["release"])
}