pipeline {
    // Forzamos ejecución en el nodo Ubuntu ARM64 (srv01)
    agent {
        label 'srv01'
    }

    environment {
        SONAR_SERVER = 'SonarQube'
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
            environment {
                SCANNER_HOME = tool 'SonarScanner'
            }
            steps {
                withSonarQubeEnv(SONAR_SERVER) {
                    sh '${SCANNER_HOME}/bin/sonar-scanner'
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
