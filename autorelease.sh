#!/usr/bin/env sh
#
# autorelease script
# created by mah0x211
#

set -eu

#
# you can generate a token at https://github.com/settings/tokens
#
#GITHUB_TOKEN='<your-githup-api-token>'

#
# install required packages
#
apt-get install jq file curl -y


#
# create binary files
#
echo "CREATE RELEASE FILES"
WORKDIR=${PWD}
for pathname in ./build/*/*; do
    echo "  create .tar.gz and checksum files for ${pathname}"
    cmdname=$(basename ${pathname})
    cmddir=$(dirname ${pathname})
    platform=$(basename ${cmddir})
    tarfile="${cmdname}-${platform}.tar.gz"
    sumfile="${tarfile}.sha256"

    cd ${cmddir}

    #
    # build .tar.gz file
    #
    echo "  tar -zcvf ${tarfile} ${cmdname}"
    tar -zcvf ${tarfile} ${cmdname}

    #
    # build .sha256 file
    #
    echo "  sha256sum ${tarfile} > ${sumfile}"
    sha256sum ${tarfile} > ${sumfile}

    #
    # remove bin file
    #
    echo "  remove ${cmdname}"
    rm ${cmdname}
    echo ""

    cd ${WORKDIR}
done


#
# repository attributes
#
echo "REPOSITORY INFO"
ORIGIN="https://${GITHUB_TOKEN}:x-oauth-basic@$(git remote get-url origin | cut -d @ -f2 | sed -e 's/:/\//')"
REPO=$(basename "${ORIGIN}" | cut -d \. -f1)
BRANCH=$(git branch | grep \* | cut -d ' ' -f2)

echo "  repository: ${REPO}"
echo "  branch    : ${BRANCH}"
echo ""


#
# add tag if not exists
#
TAG="${BRANCH}"
if [ -z "$(git tag -l | grep ^${TAG}$)" ]; then
    echo "ADD TAG: ${TAG}"
    git tag ${TAG}
    git push ${ORIGIN} --tags
    echo ""
fi


#
# create github release
#
echo "CREATE GITHUB RELEASE"
GH_OWNER='mah0x211'
GH_REPO="repos/${GH_OWNER}/${REPO}"
GH_AUTHZ_HDR="Authorization: token ${GITHUB_TOKEN}"
GH_ACCPT_HDR="Accept: application/vnd.github.v3+json"
GH_API_URL="https://api.github.com/${GH_REPO}/releases"
GH_RELEASE_JSON="{\"tag_name\": \"${TAG}\", \"name\": \"${TAG}\"}"

#
# remove current release if exists
#
RESP=$(curl -s \
        -H "${GH_AUTHZ_HDR}" \
        -H "${GH_ACCPT_HDR}" \
        ${GH_API_URL}/tags/${TAG})
RELEASE_ID=$(echo "${RESP}" | jq .id)
if [ "${RELEASE_ID}" != "null" ]; then
    # delete a release
    echo "  DELETE EXSISTING RELEASE ${RELEASE_ID}"
    curl -s \
        -H "${GH_AUTHZ_HDR}" \
        -H "${GH_ACCPT_HDR}" \
        -X DELETE ${GH_API_URL}/${RELEASE_ID}
fi

#
# create new release
#
echo "${GH_RELEASE_JSON}" | jq .
RESP=$(curl -s \
        -H "${GH_AUTHZ_HDR}" \
        -H "${GH_ACCPT_HDR}" \
        -d "${GH_RELEASE_JSON}" \
        ${GH_API_URL})
RELEASE_ID=$(echo "${RESP}" | jq .id)
# failed to create release
if [ "${RELEASE_ID}" = "null" ]; then
    echo "  failed to create release!"
    echo "${RESP}"
    exit 1
fi
echo "ok"
echo ""

#
# upload binary files
#
GH_UPLOAD_URL="https://uploads.github.com/${GH_REPO}/releases/${RELEASE_ID}/assets"
for pathname in ./build/*/*; do
    tarfile=$(basename ${pathname})
    mimetype=$(file -b --mime-type ${pathname})

    echo "  upload ${tarfile} to ${GH_UPLOAD_URL}"
    curl -s \
         -H "${GH_AUTHZ_HDR}" \
         -H "${GH_ACCPT_HDR}" \
         -H "Content-Type: ${mimetype}" \
         --data-binary "@${pathname}" \
         ${GH_UPLOAD_URL}?name=${tarfile}
    echo ""
done

