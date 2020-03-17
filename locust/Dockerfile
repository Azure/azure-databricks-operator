FROM locustio/locust:0.13.2

# locust image is user which can't install pip things so we go back to root
USER root


COPY ./locust/behaviours ./behaviours
COPY ./locust/locust_files ./locust_files

COPY locust/requirements.txt .

RUN pip install -r requirements.txt

ENTRYPOINT [ "locust" ]
