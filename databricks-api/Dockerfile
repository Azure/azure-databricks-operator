FROM continuumio/miniconda3:latest
LABEL maintainer="Azkhojan@microsoft.com"
ENV DATABRICKS_HOST ""
ENV DATABRICKS_TOKEN ""
RUN apt-get update && apt-get install gettext -y && apt-get clean
WORKDIR /tmp
COPY environment.yml ./
RUN conda env create -f environment.yml
RUN echo "source activate databricksapi" > ~/.bashrc
ENV PATH /opt/conda/envs/databricksapi/bin:$PATH
RUN /bin/bash -c "source activate databricksapi"
COPY requirements.txt ./
RUN pip install -r requirements.txt
COPY /app /app
WORKDIR /app
RUN rm -r /tmp
EXPOSE 5000
ENTRYPOINT ["bash","-c"]
CMD ["python app.py"]