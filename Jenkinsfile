pipeline {
    agent any

    environment {        
        SSH_KEY_ID = 'ark-deploy-ssh-key'        
        ENV_FILE_ID = 'ark-deploy-env-prod'
        COMPOSE_FILE_ID = 'ark-deploy-docker-compose-prod'
        SERVER_IP = 'ark-deploy-server-ip'
        PROJECT_NAME = 'ark_deploy'
    }

    stages {
        stage('1. Pre-Check & Pull') {
            steps {
                echo "Iniciando despliegue de ${env.PROJECT_NAME}..."                
            }
        }

        stage('2. Unit Tests') {
            agent {                
                docker { 
                    image 'golang:1.26-alpine'
                    args '-v /var/run/docker.sock:/var/run/docker.sock'
                }
            }
            steps {
                sh '''
                    # 1. Verificación del entorno Go
                    go version
                    
                    # 2. Instalación de dependencias
                    go mod download
                    
                    # 3. Ejecución de tests
                    echo "Ejecutando tests unitarios..."
                    go test ./... -v -cover && exit 0
                '''
            }
        }

        stage('3. Deploy to Production') {
            steps {                
                withCredentials([
                    file(credentialsId: env.ENV_FILE_ID, variable: 'ENV_FILE'),
                    file(credentialsId: env.COMPOSE_FILE_ID, variable: 'COMPOSE_FILE'),
                    string(credentialsId: env.SERVER_IP, variable: 'TARGET_IP')
                ]) {
                                    
                    ansiblePlaybook(
                        playbook: 'ci/playbook.yml',
                        inventory: 'ci/inventory.ini',
                        credentialsId: env.SSH_KEY_ID,                    
                        extraVars: [
                            env_file: "${ENV_FILE}",
                            repo_dir: "${WORKSPACE}",
                            compose_file: "${COMPOSE_FILE}",
                            ansible_host: "${TARGET_IP}"
                        ],
                        colorized: true
                    )
                }
            }
        }
    }

    post {
        success {
            echo "¡Despliegue exitoso de ${env.PROJECT_NAME}!"
        }
        failure {
            echo "xxxxx El Pipeline de ${env.PROJECT_NAME} falló. xxxxx"
        }
    }
}
