try {

    sh 'docker-compose -p ${DOCKER_COMPOSE_PROJECT_NAME} up -d etcd'
    sh 'docker-compose -p ${DOCKER_COMPOSE_PROJECT_NAME} up -d pulsar'
    dir ('build/docker/deploy') {
        sh 'docker-compose -p ${DOCKER_COMPOSE_PROJECT_NAME} pull'
        sh 'docker-compose -p ${DOCKER_COMPOSE_PROJECT_NAME} up -d'
    }

    dir ('build/docker/test') {
        sh 'docker pull ${SOURCE_REPO}/pytest:${SOURCE_TAG} || true'
        sh 'docker-compose build --force-rm regression'
        sh 'docker-compose -p ${DOCKER_COMPOSE_PROJECT_NAME} run --rm regression'
        try {
            withCredentials([usernamePassword(credentialsId: "${env.DOCKER_CREDENTIALS_ID}", usernameVariable: 'DOCKER_USERNAME', passwordVariable: 'DOCKER_PASSWORD')]) {
                sh 'docker login -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD} ${DOKCER_REGISTRY_URL}'
                sh 'docker-compose push regression'
            }
        } catch (exc) {
            throw exc
        } finally {
            sh 'docker logout ${DOKCER_REGISTRY_URL}'
        }
    }
} catch(exc) {
    throw exc
} finally {
    sh 'docker-compose -p ${DOCKER_COMPOSE_PROJECT_NAME} rm -f -s -v pulsar'
    sh 'docker-compose -p ${DOCKER_COMPOSE_PROJECT_NAME} rm -f -s -v etcd'
    dir ('build/docker/deploy') {
        sh 'docker-compose -p ${DOCKER_COMPOSE_PROJECT_NAME} down --rmi all -v || true'
    }
    dir ('build/docker/test') {
        sh 'docker-compose -p ${DOCKER_COMPOSE_PROJECT_NAME} run --rm regression /bin/bash -c "rm -rf __pycache__ && rm -rf .pytest_cache"'
        sh 'docker-compose -p ${DOCKER_COMPOSE_PROJECT_NAME} down --rmi all -v || true'
    }
}