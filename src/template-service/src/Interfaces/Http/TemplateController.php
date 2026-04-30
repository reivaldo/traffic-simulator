<?php

declare(strict_types=1);

namespace TemplateService\Interfaces\Http;

use TemplateService\Application\RenderTemplate;
use TemplateService\Domain\TemplateRequest;

final class TemplateController
{
    private RenderTemplate $useCase;

    public function __construct(RenderTemplate $useCase)
    {
        $this->useCase = $useCase;
    }

    /**
     * @return array{status:int, body:array<string,mixed>}
     */
    public function send(string $rawBody): array
    {
        /** @var mixed $payload */
        $payload = json_decode($rawBody, true);
        if (!is_array($payload)) {
            return [
                'status' => 400,
                'body' => ['status' => 'error', 'error' => 'invalid json'],
            ];
        }

        try {
            $request = TemplateRequest::fromArray($payload);
            $result = $this->useCase->execute($request);

            return [
                'status' => 200,
                'body' => $result,
            ];
        } catch (\InvalidArgumentException $e) {
            return [
                'status' => 400,
                'body' => ['status' => 'error', 'error' => $e->getMessage()],
            ];
        } catch (\Throwable $e) {
            return [
                'status' => 500,
                'body' => ['status' => 'error', 'error' => 'internal error'],
            ];
        }
    }
}

