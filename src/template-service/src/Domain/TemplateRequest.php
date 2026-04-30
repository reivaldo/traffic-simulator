<?php

declare(strict_types=1);

namespace TemplateService\Domain;

final class TemplateRequest
{
    private string $externalId;
    private string $recipient;
    private string $channel;

    private function __construct(string $externalId, string $recipient, string $channel)
    {
        $this->externalId = $externalId;
        $this->recipient = $recipient;
        $this->channel = $channel;
    }

    /**
     * @param array<string,mixed> $data
     */
    public static function fromArray(array $data): self
    {
        $externalId = trim((string) ($data['external_id'] ?? ''));
        $recipient = trim((string) ($data['to'] ?? ''));
        $channel = trim((string) ($data['channel'] ?? ''));

        if ($externalId === '') {
            throw new \InvalidArgumentException('external_id is required');
        }
        if ($recipient === '') {
            throw new \InvalidArgumentException('to is required');
        }
        if ($channel === '') {
            throw new \InvalidArgumentException('channel is required');
        }

        return new self($externalId, $recipient, strtolower($channel));
    }

    public function externalId(): string
    {
        return $this->externalId;
    }

    public function recipient(): string
    {
        return $this->recipient;
    }

    public function channel(): string
    {
        return $this->channel;
    }
}

