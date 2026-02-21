pipeline {
  agent any

  environment {
    SSH_KEY_ID = 'ark-deploy-ssh-key'
    ENV_FILE_ID = 'ark-deploy-env-prod'
    COMPOSE_FILE_ID = 'ark-deploy-docker-compose-prod'
    SERVER_IP = 'ark-deploy-server-ip'
    PROJECT_NAME = 'ark_deploy'

    GHCR_TOKEN_ID = 'GHCR_TOKEN'
    GHCR_USER = 'raztreuzz'
    ARK_BACKEND_IMAGE = 'ghcr.io/raztreuzz/ark_deploy-backend:prod'
    ARK_FRONTEND_IMAGE = 'ghcr.io/raztreuzz/ark_deploy-frontend:prod'

    ANSIBLE_HOST_KEY_CHECKING = 'False'
  }

  stages {
    stage('1. Unit Tests (Go)') {
      agent { docker { image 'golang:1.26' } }
      steps {
        sh '''
          set -e
          go version
          go mod download
          PKGS=$(go list ./... | grep -v '/cmd/test_api' | grep -v '/internal/tailscale')
          go test $PKGS -v
        '''
      }
    }

    stage('2. Build & Push Images') {
      steps {
        withCredentials([string(credentialsId: env.GHCR_TOKEN_ID, variable: 'GHCR_TOKEN')]) {
          sh '''
            set -e
            echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USER" --password-stdin

            docker build -t "$ARK_BACKEND_IMAGE" .
            docker push "$ARK_BACKEND_IMAGE"

            if [ -d "./frontend" ]; then
              docker build -t "$ARK_FRONTEND_IMAGE" ./frontend
              docker push "$ARK_FRONTEND_IMAGE"
            else
              echo "No existe ./frontend, saltando build/push de frontend."
            fi
          '''
        }
      }
    }

    stage('3. Deploy to Production') {
      steps {
        echo "Desplegando ${env.PROJECT_NAME} a producción..."

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
              env_file: ENV_FILE,
              compose_file: COMPOSE_FILE,
              ansible_host: TARGET_IP
            ],
            colorized: true
          )
        }
      }
    }
  }

  post {
    success { echo "¡Despliegue exitoso de ${env.PROJECT_NAME}!" }
    failure { echo "xxxxx El Pipeline de ${env.PROJECT_NAME} falló. xxxxx" }
  }
}