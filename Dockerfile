FROM scratch AS machine_manager
EXPOSE 11111
COPY machine_manager_bin /
ENTRYPOINT ["/machine_manager_bin"]

FROM scratch AS loadbalancer
EXPOSE 22222 33333
COPY loadbalancer_bin /
ENTRYPOINT ["/loadbalancer_bin", "--address=localhost", "--data-port=22222", "--admin-port=33333"]

FROM scratch AS python_linter
EXPOSE 44444
COPY python_linter_bin /
ENTRYPOINT ["/python_linter_bin", "--address=localhost", "--port=44444"]

