<?php

declare(strict_types=1);

namespace TemplateService\Tests;

use PHPUnit\Framework\TestCase;
use TemplateService\Application\RenderTemplate;
use TemplateService\Infrastructure\Renderer\SimpleTemplateRenderer;
use TemplateService\Interfaces\Http\TemplateController;

final class TemplateControllerTest extends TestCase
{
    public function testSendReturns200ForValidPayload(): void
    {
        $controller = new TemplateController(
            new RenderTemplate(new SimpleTemplateRenderer())
        );

        $response = $controller->send((string) json_encode([
            'external_id' => 'e-1',
            'to' => 'person@example.com',
            'channel' => 'email',
        ], JSON_THROW_ON_ERROR));

        self::assertSame(200, $response['status']);
        self::assertSame('success', $response['body']['status']);
    }

    public function testSendRejectsInvalidJson(): void
    {
        $controller = new TemplateController(
            new RenderTemplate(new SimpleTemplateRenderer())
        );
        $response = $controller->send('{invalid-json');

        self::assertSame(400, $response['status']);
        self::assertSame('error', $response['body']['status']);
    }
}

