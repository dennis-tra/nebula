FROM python:3.10.0

WORKDIR /app

RUN pip install poetry==1.2.0

COPY pyproject.toml poetry.lock ./
RUN poetry install

COPY . .

ENTRYPOINT ./entrypoint.sh