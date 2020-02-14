local circle = import 'circle.libsonnet';

circle.ServiceConfig('twilight') {
  jobs+: {
    tests: circle.Job(dockerImage='circleci/golang:1.13', withDocker=false) {
      steps_+:: [
        circle.RestoreCacheStep('go-deps-{{ checksum "go.sum" }}'),
        circle.RunStep('Fetch Dependencies', 'go mod vendor'),
        circle.SaveCacheStep('go-deps-{{ checksum "go.sum" }}', ['vendor']),
        circle.RunStep('Run Tests', 'make test')
      ],
    },
  },
  workflows+: {
    ['build-push']+: {
      jobs_:: [
        'tests', 
        {
          name:: 'build',
          requires: ['tests'],
        }
      ],
    },
  },
}