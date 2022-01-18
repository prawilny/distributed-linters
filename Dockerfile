FROM scratch AS machine_manager
COPY machine_manager_bin /
ENTRYPOINT ["/machine_manager_bin"]

FROM scratch AS loadbalancer
COPY loadbalancer_bin /
ENTRYPOINT ["/loadbalancer_bin"]

FROM scratch AS python_linter
COPY python_linter_bin /
ENTRYPOINT ["/python_linter_bin"]

