<?php

declare(strict_types=1);

namespace TemplateService\Tests;

use PHPUnit\Framework\TestCase;
use TemplateService\Application\RenderTemplate;
use TemplateService\Domain\TemplateRequest;
use TemplateService\Infrastructure\Renderer\SimpleTemplateRenderer;

final class RenderTemplateTest extends TestCase
{
    public function testExecuteUsesChannelTemplate(): void
    {
        $useCase = new RenderTemplate(new SimpleTemplateRenderer());
        $request = TemplateRequest::fromArray([
            'external_id' => 'ext-22',
            'to' => '+351900000000',
            'channel' => 'sms',
        ]);

        $result = $useCase->execute($request);

        self::assertSame('success', $result['status']);
        self::assertSame('sms-default-v1', $result['template_id']);
        self::assertStringContainsString('[SMS]', $result['rendered_message']);
        self::assertStringContainsString('ext-22', $result['rendered_message']);
    }
}

