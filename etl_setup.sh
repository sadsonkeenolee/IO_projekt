#!/bin/bash

readonly DIR="temp"
readonly GOODREADS_BOOKS_DATA="$(pwd)/$DIR/goodreads-books-data.zip"
readonly GOODREADS_BOOKS_DIR="$(pwd)/$DIR/goodreads-books-data/"
readonly COMMON_BOOKS_DATA="$(pwd)/$DIR/books-data.zip"
readonly COMMON_BOOKS_DIR="$(pwd)/$DIR/books-data/"
readonly TMDB_MOVIES_DATA="$(pwd)/$DIR/tmdb-movies-data.zip"
readonly TMDB_MOVIES_DIR="$(pwd)/$DIR/tmdb-movies-data/"
readonly COMMON_MOVIES_DATA="$(pwd)/$DIR/movies-data.zip"
readonly COMMON_MOVIES_DIR="$(pwd)/$DIR/movies-data/"
mkdir -p $(pwd)/$DIR

if [ ! -e "$GOODREADS_BOOKS_DATA" ]; then
  curl -L -o "$GOODREADS_BOOKS_DATA" https://www.kaggle.com/api/v1/datasets/download/jealousleopard/goodreadsbooks &
fi

if [ ! -e "$COMMON_BOOKS_DATA" ]; then
  curl -L -o "$COMMON_BOOKS_DATA" https://www.kaggle.com/api/v1/datasets/download/elvinrustam/books-dataset &
fi

if [ ! -e "$TMDB_MOVIES_DATA" ]; then
  curl -L -o "$TMDB_MOVIES_DATA" https://www.kaggle.com/api/v1/datasets/download/tmdb/tmdb-movie-metadata &
fi

if [ ! -e "$COMMON_MOVIES_DATA" ]; then
  curl -L -o "$COMMON_MOVIES_DATA" https://www.kaggle.com/api/v1/datasets/download/rounakbanik/the-movies-dataset &
fi
wait

unzip "$GOODREADS_BOOKS_DATA" -d "$GOODREADS_BOOKS_DIR"
unzip "$COMMON_BOOKS_DATA" -d "$COMMON_BOOKS_DIR"
unzip "$TMDB_MOVIES_DATA" -d "$TMDB_MOVIES_DIR"
unzip "$COMMON_MOVIES_DATA" -d "$COMMON_MOVIES_DIR"

rm $(pwd)/$DIR/*.zip
