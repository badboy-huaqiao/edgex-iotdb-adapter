docker run -d -p 59990:59990\
    -e DATABASES_PRIMARY_HOST=iotdb_host \
    -e MESSAGEQUEUE_HOST=edgex-redis \
    --network edgex-network \
    --name edgex-iotdb-adapter \
    huaqiaoz/edgex-iotdb-adapter:0.1.0

docker run -d -p 59990:59990\
    -e MESSAGEQUEUE_HOST=edgex-redis \
    --network edgex-compose_edgex-network \
    --name edgex-iotdb-adapter \
    huaqiaoz/edgex-iotdb-adapter:0.1.0