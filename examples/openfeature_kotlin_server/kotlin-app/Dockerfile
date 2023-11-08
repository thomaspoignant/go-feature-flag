# Use an official Kotlin image as the base image
FROM openjdk:11-jre-slim

# Set the working directory in the container
WORKDIR /app

COPY . /app
RUN ./gradlew shadowJar

# Command to run your application
CMD ["java", "-jar", "/app/build/libs/openfeature_kotlin-1.0-SNAPSHOT-all.jar"]