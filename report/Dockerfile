FROM python:3.10-bullseye
WORKDIR /app

RUN apt-get update
RUN apt-get install -y markdown
RUN pip install poetry==1.2.0

COPY pyproject.toml poetry.lock ./
RUN poetry install

COPY . .
RUN ["chmod", "+x", "/app/entrypoint.sh"]

ENTRYPOINT ./entrypoint.sh
