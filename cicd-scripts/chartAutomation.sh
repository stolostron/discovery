# Copyright (c) 2020 Red Hat, Inc.

#Script to be run in travis when charts are updated in master

cicd-scripts/install-dependencies.sh
cicd-scripts/chart-sync.sh
git add .
git commit -m "Update from $1"
git push https://${MCH_REPO_BOT_TOKEN}@github.com/open-cluster-management/multicloudhub-repo.git "master"
