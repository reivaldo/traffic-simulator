<?php

declare(strict_types=1);

namespace TemplateService\Application\Ports;

use TemplateService\Domain\TemplateRequest;

interface TemplateRendererPort
{
    /**
     * @return array{template_id:string, message:string}
     */
    public function render(TemplateRequest $request): array;
}

