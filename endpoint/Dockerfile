ARG IMAGE=scratch
FROM $IMAGE

LABEL MAINTAINER="bitschen@163.com"

# 执行文件名称，须与 name.txt 中一致
COPY endpoint-dongkong /bin/

ONBUILD COPY application.toml /etc/edgex/

CMD ["endpoint-dongkong"]