<?php

declare(strict_types=1);

namespace TemplateService\Infrastructure\Renderer;

use TemplateService\Application\Ports\TemplateRendererPort;
use TemplateService\Domain\TemplateRequest;

final class SimpleTemplateRenderer implements TemplateRendererPort
{
    /**
     * @var array<string,string>
     */
    private array $templates = [
        'sms' => '[SMS] Message for {{recipient}} ({{external_id}})',
        'email' => '[EMAIL] Message for {{recipient}} ({{external_id}})',
        'whatsapp' => '[WHATSAPP] Message for {{recipient}} ({{external_id}})',
    ];

    public function render(TemplateRequest $request): array
    {
        $templateId = 'default';
        $template = $this->templates[$request->channel()] ?? '[GENERIC] Message for {{recipient}} ({{external_id}})';

        if (isset($this->templates[$request->channel()])) {
            $templateId = $request->channel() . '-default-v1';
        }

        $message = str_replace(
            ['{{recipient}}', '{{external_id}}'],
            [$request->recipient(), $request->externalId()],
            $template
        );

        return [
            'template_id' => $templateId,
            'message' => $message,
        ];
    }
}

