# Use the official PostgreSQL image as base
FROM postgres:16

# Set environment variables
ENV POSTGRES_DB=chirpy
ENV POSTGRES_USER=user
ENV POSTGRES_PASSWORD=postgres

# Expose the default PostgreSQL port
EXPOSE 5432
