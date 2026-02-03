plugins {
    id("com.android.library")
    id("org.jetbrains.kotlin.android")
    id("org.jetbrains.kotlin.plugin.serialization")
    id("signing")
    id("maven-publish")
    id("org.jlleitschuh.gradle.ktlint")
}

val releaseVersion = project.extra["version"].toString()

android {
    namespace = "org.gofeatureflag.openfeature"
    compileSdk = 35

    testOptions {
        unitTests {
            isIncludeAndroidResources = true
            all { it.useJUnit() }
        }
    }

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
    api("dev.openfeature:kotlin-sdk:${rootProject.extra["kotlinSdkVersion"]}")
    api("io.ktor:ktor-client-okhttp:${rootProject.extra["ktorVersion"]}")
    api("io.ktor:ktor-client-content-negotiation:${rootProject.extra["ktorVersion"]}")
    api("io.ktor:ktor-serialization-kotlinx-json:${rootProject.extra["ktorVersion"]}")
    api("com.squareup.okhttp3:okhttp:${rootProject.extra["okhttpVersion"]}")
    api("com.google.code.gson:gson:${rootProject.extra["gsonVersion"]}")
    api("org.jetbrains.kotlinx:kotlinx-coroutines-core:${rootProject.extra["kotlinxCoroutinesCoreVersion"]}")
    api("org.jetbrains.kotlinx:kotlinx-serialization-json:${rootProject.extra["kotlinxSerializationJsonVersion"]}")
    testImplementation("junit:junit:${rootProject.extra["junitVersion"]}")
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:${rootProject.extra["kotlinxCoroutinesTestVersion"]}")
    testImplementation("com.squareup.okhttp3:mockwebserver:${rootProject.extra["okhttpVersion"]}")
    testImplementation("org.skyscreamer:jsonassert:${rootProject.extra["jsonassertVersion"]}")
    testImplementation("io.ktor:ktor-client-mock:${rootProject.extra["ktorClientMockVersion"]}")
    testImplementation(kotlin("test"))
    testImplementation("org.robolectric:robolectric:${rootProject.extra["robolectricVersion"]}")
    testImplementation("org.slf4j:slf4j-simple:${rootProject.extra["slf4jSimpleVersion"]}")
}

signing {
    sign(publishing.publications["release"])
}