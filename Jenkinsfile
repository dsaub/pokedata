pipeline {
    // Forzamos ejecución en el nodo Ubuntu ARM64 (srv01)
    agent {
        label 'srv01'
    }

    environment {
        SONAR_SERVER = 'main'
        // Directorios estándar de caché de Go en srv01
        GOPATH       = "${WORKSPACE}/.go"
        GOCACHE      = "${WORKSPACE}/.cache/go-build"
        PATH         = "/usr/local/go/bin:${WORKSPACE}/.go/bin:${env.PATH}"
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Test & Coverage') {
            steps {
                echo 'Ejecutando tests unitarios y generando informe de cobertura...'
                sh '''
                    # Ejecuta los tests generando coverage.out
                    go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
                '''
            }
        }

        stage('SonarQube Analysis') {
            def scannerHome = tool 'SonarScanner';
            withSonarQubeEnv() {
                sh "${scannerHome}/bin/sonar-scanner"
            }
        }

        stage('Quality Gate') {
            steps {
                timeout(time: 2, unit: 'MINUTES') {
                    // Espera a que SonarQube procese el reporte y valide si pasa el gate
                    waitForQualityGate abortPipeline: true
                }
            }
        }
    }

    post {
        always {
            cleanWs(deleteDirs: true, notFailBuild: true)
        }
        success {
            echo '¡Build, tests y análisis de SonarQube completados exitosamente!'
        }
    }
}