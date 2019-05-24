ARG BUSYBOX_IMG=busybox
FROM $BUSYBOX_IMG

LABEL MAINTAINER="yoojiachen@gmail.com"

# 执行文件名称，须与 name.txt 中一致
COPY edgex-trigger-dongkong /bin/

ONBUILD COPY application.toml /etc/edgex/

CMD ["edgex-trigger-dongkong"]