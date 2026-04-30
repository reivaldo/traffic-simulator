<?php

declare(strict_types=1);

namespace TemplateService\Interfaces\Http;

final class Router
{
    private TemplateController $controller;

    public function __construct(TemplateController $controller)
    {
        $this->controller = $controller;
    }

    public function handle(string $method, string $path, string $body): void
    {
        if ($path === '/healthz' || $path === '/readyz') {
            $this->json(200, ['status' => 'ok']);
            return;
        }

        if ($path === '/send' && strtoupper($method) === 'POST') {
            $response = $this->controller->send($body);
            $this->json($response['status'], $response['body']);
            return;
        }

        if ($path === '/') {
            $this->json(200, [
                'service' => 'template-service',
                'language' => 'php',
                'endpoints' => ['/healthz', '/readyz', '/send'],
            ]);
            return;
        }

        $this->json(404, ['status' => 'error', 'error' => 'not found']);
    }

    /**
     * @param array<string,mixed> $body
     */
    private function json(int $status, array $body): void
    {
        http_response_code($status);
        header('Content-Type: application/json');
        echo json_encode($body, JSON_UNESCAPED_SLASHES);
    }
}

