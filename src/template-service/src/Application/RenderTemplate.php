<?php

declare(strict_types=1);

namespace TemplateService\Application;

use TemplateService\Application\Ports\TemplateRendererPort;
use TemplateService\Domain\TemplateRequest;

final class RenderTemplate
{
    private TemplateRendererPort $renderer;

    public function __construct(TemplateRendererPort $renderer)
    {
        $this->renderer = $renderer;
    }

    /**
     * @return array{
     *   status:string,
     *   template_id:string,
     *   external_id:string,
     *   channel:string,
     *   recipient:string,
     *   rendered_message:string
     * }
     */
    public function execute(TemplateRequest $request): array
    {
        $rendered = $this->renderer->render($request);

        return [
            'status' => 'success',
            'template_id' => $rendered['template_id'],
            'external_id' => $request->externalId(),
            'channel' => $request->channel(),
            'recipient' => $request->recipient(),
            'rendered_message' => $rendered['message'],
        ];
    }
}

