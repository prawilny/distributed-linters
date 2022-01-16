FROM scratch AS machine_manager
EXPOSE 11111
COPY machine_manager_bin /
ENTRYPOINT ["/machine_manager_bin", "-address=0.0.0.0:2137"]

FROM scratch AS loadbalancer
EXPOSE 22222 33333
COPY loadbalancer_bin /
ENTRYPOINT ["/loadbalancer_bin", "-data-port=22222", "-admin-port=33333"]

FROM scratch AS python_linter
EXPOSE 44444 55555
COPY python_linter_bin /
ENTRYPOINT ["/python_linter_bin", "-http-port=44444", "-grpc-port=55555"]

