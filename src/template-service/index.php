<?php

declare(strict_types=1);

require_once __DIR__ . '/vendor/autoload.php';

use TemplateService\Application\RenderTemplate;
use TemplateService\Infrastructure\Renderer\SimpleTemplateRenderer;
use TemplateService\Interfaces\Http\Router;
use TemplateService\Interfaces\Http\TemplateController;

$renderer = new SimpleTemplateRenderer();
$useCase = new RenderTemplate($renderer);
$controller = new TemplateController($useCase);
$router = new Router($controller);

$method = $_SERVER['REQUEST_METHOD'] ?? 'GET';
$path = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH);
if (!is_string($path) || $path === '') {
    $path = '/';
}
$body = file_get_contents('php://input');
if (!is_string($body)) {
    $body = '';
}

$router->handle($method, $path, $body);
