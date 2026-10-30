plugins {
    application
}

repositories {
    mavenCentral()
}

dependencies {
    testImplementation("junit:junit:4.13.2")
    testImplementation("org.hamcrest:hamcrest:2.2")
    testImplementation("de.bechte.junit:junit-hierarchicalcontextrunner:4.12.2")
}

tasks.withType<JavaCompile> {
    options.release = 11
}

// Keep the original (non-Maven) source layout.
sourceSets {
    main {
        java.setSrcDirs(listOf("src"))
    }
    test {
        java.setSrcDirs(listOf("test"))
    }
}

application {
    mainClass = "smc.SMC"
}

tasks.jar {
    archiveFileName = "smc.jar"
    // test_cases/java_turnstile/build.xml expects the jar at build/jar/smc.jar.
    destinationDirectory = layout.buildDirectory.dir("jar")
    manifest {
        attributes("Main-Class" to application.mainClass)
    }
}
