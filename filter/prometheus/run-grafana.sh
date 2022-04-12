docker run -d \
--name grafana \
-v /home/jzd/GolandProjects/bee-micro/filter/prometheus/grafana.ini:/etc/grafana/grafana.ini \
-p 65030:3000 \
grafana/grafana
