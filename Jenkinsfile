pipeline {
    environment {
        REGISTRY = 'registry-intl.ap-southeast-1.aliyuncs.com/swmeng/ddns-'
        registryCredential = 'aliclouddocker'
        DOCKER_CREDENTIALS = credentials('aliclouddocker')

    }
    agent any
    stages {
        stage('Build and Push Images') {
            steps {
                script {
                    docker.withRegistry('https://registry-intl.ap-southeast-1.aliyuncs.com', registryCredential ) {
                        sh "docker buildx create --name mybuilder --use --bootstrap"
                        sh "docker buildx build --push --platform linux/arm64,linux/amd64 -t ${REGISTRY}client:latest ./client"
                        sh "docker buildx build --push --platform linux/arm64,linux/amd64 -t ${REGISTRY}server:latest ./server"
                    }
                }
            }
        }
    }
}
