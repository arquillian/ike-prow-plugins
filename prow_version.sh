#!/usr/bin/env bash
grep "k8s.io/test-infra" ${PWD}/glide.yaml -A 1 | grep version | awk '{print $$2}'