FROM php:8.2-cli
WORKDIR /var/www/html

COPY composer.json composer.lock* phpunit.xml ./
COPY src ./src
RUN php -r "copy('https://getcomposer.org/installer', 'composer-setup.php');" \
    && php composer-setup.php --install-dir=/usr/local/bin --filename=composer \
    && rm composer-setup.php \
    && composer install --no-dev --optimize-autoloader --no-interaction --prefer-dist

COPY index.php ./
EXPOSE 8085
CMD ["php", "-S", "0.0.0.0:8085", "index.php"]
