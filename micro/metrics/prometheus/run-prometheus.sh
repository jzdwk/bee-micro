docker run -d \
-p 65090:9090 \
-v /home/jzd/GolandProjects/bee-micro/micro/metrics/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml \
--name prometheus prom/prometheus
