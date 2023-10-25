pipeline {
    environment {
        REGISTRY = 'registry-intl.ap-southeast-1.aliyuncs.com/swmeng/ddns-'
        registryCredential = 'aliclouddocker'
        DOCKER_CREDENTIALS = credentials('aliclouddocker')
        DOCKER_BUILDKIT = '1'
    }
    agent any
    stages {
        stage('Build and Push Images') {
            steps {
                script {
                    docker.withRegistry('https://registry-intl.ap-southeast-1.aliyuncs.com', registryCredential ) {
                        sh "docker buildx build --push --platform linux/arm64,linux/amd64 -t ${env.REGISTRY}client:latest ./client"
                        sh "docker buildx build --push --platform linux/arm64,linux/amd64 -t ${env.REGISTRY}server:latest ./server"
                    }
                }
            }
        }
        stage('Remove Unused Docker Image') {
            steps {
                sh "docker rmi ${registry}client"
                sh "docker rmi ${registry}server"
            }
        }
    }
}
