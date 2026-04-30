<?php

declare(strict_types=1);

namespace TemplateService\Tests;

use InvalidArgumentException;
use PHPUnit\Framework\TestCase;
use TemplateService\Domain\TemplateRequest;

final class TemplateRequestTest extends TestCase
{
    public function testFromArrayValidatesAndNormalizesInput(): void
    {
        $request = TemplateRequest::fromArray([
            'external_id' => 'msg-1',
            'to' => 'user@example.com',
            'channel' => 'EMAIL',
        ]);

        self::assertSame('msg-1', $request->externalId());
        self::assertSame('user@example.com', $request->recipient());
        self::assertSame('email', $request->channel());
    }

    public function testFromArrayThrowsWhenRequiredFieldsAreMissing(): void
    {
        $this->expectException(InvalidArgumentException::class);
        TemplateRequest::fromArray(['channel' => 'sms']);
    }
}

