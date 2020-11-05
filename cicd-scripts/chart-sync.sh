# Copyright (c) 2020 Red Hat, Inc.

cd $(dirname $0)
CHARTS_PATH="../../../../../multiclusterhub/charts"
CICD_FOLDER="../../../../"
CHART_VERSION="$(cat ../CHART_VERSION)"
echo "Fetching charts from csv"
rm ../multiclusterhub/charts/* #rm all charts first, in case chart versions are changed
while IFS=, read -r f1 f2
do
  mkdir -p tmp
  cd tmp
  #if this is being run by chart travis, use that token in travis job to clone
  if [ -z "${TRAVIS_BUILD_DIR}" ] || [[ "$(echo $TRAVIS_BUILD_DIR | cut -f8 -d/)" == "multicloudhub-repo" ]]; then 
    git clone $f1
  else 
    git clone "https://${MCH_REPO_BOT_TOKEN}@github.com/open-cluster-management/$(echo $f1 | cut -f2 -d/)"
  fi
  var1=$(echo "$(ls)" | cut -f5 -d/)  #get the repo name
  cd */ 
  git checkout $f2
  var2=$(echo $var1 | cut -f1 -d-) #get the first word (ie kui in kui-web-terminal)
  var3=$(find . -type d -name "$var2*") #look for folder in repo that starts with ^ rather than stable/*/
  cd $var3
  PACKAGE="$(helm package ./ --version $CHART_VERSION)"
  find . -type f -name "*tgz" | xargs -I '{}' mv '{}' $CHARTS_PATH
  cd $CICD_FOLDER
  rm -rf tmp
done < chartSHA.csv
helm repo index --url http://multiclusterhub-repo:3000/charts ../multiclusterhub/charts