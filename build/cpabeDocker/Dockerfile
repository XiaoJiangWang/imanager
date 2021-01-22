FROM 10.5.26.86:8080/zjlab/centos:centos7

ENV APP_DIR /home/zjlab

RUN mkdir -p ${APP_DIR} \
&& yum -y install m4 \
&& yum -y install flex \
&& yum -y install bison \
&& yum -y install wget \
&& wget https://ftp.tu-chemnitz.de/pub/linux/dag/redhat/el7/en/x86_64/rpmforge/RPMS/lzip-1.15-1.el7.rf.x86_64.rpm \
&& rpm -i lzip*rpm \
&& yum -y install gcc gcc-c++ kernel-devel make \
&& yum -y install gmp-devel \
&& rm lzip-1.15-1.el7.rf.x86_64.rpm \
&& wget https://crypto.stanford.edu/pbc/files/pbc-0.5.14.tar.gz \
&& tar -xvf pbc-0.5.14.tar.gz \
&& rm pbc-0.5.14.tar.gz \
&& cd pbc-0.5.14 \
&& yum -y install m4 \
&& yum -y install flex \
&& yum -y install bison \
&& ./configure \
&& make \
&& make install \
&& cd .. \
&& yum -y install openssl-devel \
&& yum -y install glib2-devel.x86_64 \
&& wget http://acsc.cs.utexas.edu/cpabe/libbswabe-0.9.tar.gz \
&& tar -xvf libbswabe-0.9.tar.gz \
&& rm libbswabe-0.9.tar.gz \
&& cd libbswabe-0.9 \
&& ./configure \
&& make \
&& make install \
&& cd .. \
&& wget http://acsc.cs.utexas.edu/cpabe/cpabe-0.11.tar.gz \
&& tar -xvf cpabe-0.11.tar.gz \
&& rm cpabe-0.11.tar.gz \
&& cd cpabe-0.11 \
&& ./configure \
&& cat Makefile \
&& sed -i 's/-lcrypto -lcrypto/-lcrypto -lcrypto \\\n\t-lgmp/' Makefile \
&& sed -i '67s/}/;}/' policy_lang.y \
&& make \
&& make install \
&& cd .. \
&& cpabe-setup