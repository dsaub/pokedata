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
            steps {
                withSonarQubeEnv("${SONAR_SERVER}") {
                    sh '''
                        sonar-scanner \
                          -Dsonar.projectKey=pokedata \
                          -Dsonar.projectName="PokeDATA" \
                          -Dsonar.sources=. \
                          -Dsonar.exclusions="**/*_test.go,**/vendor/**,.go/**,.cache/**" \
                          -Dsonar.tests=. \
                          -Dsonar.test.inclusions="**/*_test.go" \
                          -Dsonar.test.exclusions=".go/**,.cache/**" \
                          -Dsonar.go.coverage.reportPaths=coverage.out
                    '''
                }
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